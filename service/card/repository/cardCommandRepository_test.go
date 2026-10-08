package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
)

// stubConfig scripts the behavior of the in-process driver stub. On the
// happy path (all errors nil) the returned rows are zero-valued — the tests
// only assert on the returned error, not on row contents.
//
// The card mutation queries (ToggleCardStatus, UpdateCreditLimit,
// RedeemRewardPoints) are UPDATE statements scanned through RETURNING, while
// the GetCardByID fallback used by RedeemPoints is a SELECT — the stub
// dispatches on that difference.
type stubConfig struct {
	queryRowErr error // error returned by the card mutation query
	getCardErr  error // error returned by the GetCardByID fallback query
	getCardOK   bool  // when true, GetCardByID returns a (zero-valued) row
}

// stubDriver is a minimal database/sql driver whose connections answer every
// query from the scripted stubConfig, so the repository error paths can be
// exercised without a live database.
type stubDriver struct {
	cfg *stubConfig
}

type stubConn struct {
	cfg *stubConfig
}

func (d stubDriver) Open(string) (driver.Conn, error) { return stubConn{cfg: d.cfg}, nil }

func (c stubConn) Prepare(query string) (driver.Stmt, error) {
	return stubStmt{cfg: c.cfg, query: query}, nil
}
func (c stubConn) Close() error              { return nil }
func (c stubConn) Begin() (driver.Tx, error) { return nil, errors.New("not implemented") }

type stubStmt struct {
	cfg   *stubConfig
	query string
}

func (s stubStmt) Close() error  { return nil }
func (s stubStmt) NumInput() int { return -1 }

func (s stubStmt) Exec([]driver.Value) (driver.Result, error) {
	return driver.Result(driver.RowsAffected(0)), nil
}

// cardBaseColumns covers every field of GetCardByIDRow / UpdateCardRow.
var cardBaseColumns = []string{
	"card_id", "user_id", "card_number", "card_type", "expire_date", "cvv",
	"card_provider", "created_at", "updated_at",
}

// cardStatusColumns extends the base set for the RETURNING clause of the
// status/credit-limit/redeem mutations (ToggleCardStatusRow et al.).
var cardStatusColumns = append(append([]string{}, cardBaseColumns...),
	"status", "credit_limit", "outstanding_balance", "reward_points")

// cardDeletedColumns extends the base set for trash/restore RETURNING
// (TrashCardRow / RestoreCardRow).
var cardDeletedColumns = append(append([]string{}, cardBaseColumns...), "deleted_at")

func mutationColumns(query string) []string {
	switch {
	case strings.Contains(query, "SET status = CASE WHEN status"):
		return cardStatusColumns
	case strings.Contains(query, "SET credit_limit = $2"):
		return cardStatusColumns
	case strings.Contains(query, "reward_points = GREATEST"):
		return cardStatusColumns
	case strings.Contains(query, "deleted_at = current_timestamp"):
		return cardDeletedColumns
	case strings.Contains(query, "deleted_at = NULL"):
		return cardDeletedColumns
	default:
		return cardBaseColumns
	}
}

func (s stubStmt) Query([]driver.Value) (driver.Rows, error) {
	if strings.HasPrefix(strings.TrimSpace(s.query), "SELECT") {
		// Only GetCardByID reaches the repository as a plain SELECT.
		if s.cfg.getCardOK {
			return &stubRows{cols: cardBaseColumns}, nil
		}
		return nil, s.cfg.getCardErr
	}
	return &stubRows{cols: mutationColumns(s.query)}, s.cfg.queryRowErr
}

// stubRows yields a single row of zero values (NULL for timestamp-like
// columns, which pgtype types accept) and then EOF.
type stubRows struct {
	cols     []string
	consumed bool
}

func (r *stubRows) Columns() []string { return r.cols }
func (r *stubRows) Close() error      { return nil }

func (r *stubRows) Next(dest []driver.Value) error {
	if r.consumed {
		return io.EOF
	}
	r.consumed = true
	for i, col := range r.cols {
		switch col {
		case "created_at", "updated_at", "deleted_at", "expire_date":
			dest[i] = nil
		case "card_id", "user_id", "credit_limit", "outstanding_balance", "reward_points":
			dest[i] = int64(0)
		default:
			dest[i] = ""
		}
	}
	return nil
}

var stubDriverSeq int32

// newFakeCardCommandRepository wires a cardCommandRepository to a fresh
// in-process sqlx.DB backed by the stub driver with the given script.
func newFakeCardCommandRepository(t *testing.T, cfg *stubConfig) CardCommandRepository {
	t.Helper()

	name := fmt.Sprintf("cardCommandStub%d", atomic.AddInt32(&stubDriverSeq, 1))
	sql.Register(name, stubDriver{cfg: cfg})

	conn, err := sqlx.Connect(name, "stub://card-command")
	if err != nil {
		t.Fatalf("sqlx.Connect: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return NewCardCommandRepository(conn)
}

// assertAppErrorCode verifies the returned error is an AppError with the
// expected HTTP code and message.
func assertAppErrorCode(t *testing.T, err error, wantCode int, wantMsg string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var appErr *sharedErrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *sharedErrors.AppError, got %T: %v", err, err)
	}
	if appErr.Code != wantCode {
		t.Fatalf("AppError.Code = %d, want %d (message: %s)", appErr.Code, wantCode, appErr.Message)
	}
	if appErr.Message != wantMsg {
		t.Fatalf("AppError.Message = %q, want %q", appErr.Message, wantMsg)
	}
}

func TestToggleCardStatusCardNotFound(t *testing.T) {
	repo := newFakeCardCommandRepository(t, &stubConfig{
		queryRowErr: sql.ErrNoRows,
	})

	_, err := repo.ToggleCardStatus(context.Background(), &requests.ToggleCardStatusRequest{CardID: 99})

	assertAppErrorCode(t, err, http.StatusNotFound, "card not found")
}
func TestToggleCardStatusInternalError(t *testing.T) {
	repo := newFakeCardCommandRepository(t, &stubConfig{
		queryRowErr: errors.New("connection reset"),
	})

	_, err := repo.ToggleCardStatus(context.Background(), &requests.ToggleCardStatusRequest{CardID: 1})

	assertAppErrorCode(t, err, http.StatusInternalServerError, "Failed to update card status")
}

func TestToggleCardStatusSuccess(t *testing.T) {
	repo := newFakeCardCommandRepository(t, &stubConfig{})

	if _, err := repo.ToggleCardStatus(context.Background(), &requests.ToggleCardStatusRequest{CardID: 1}); err != nil {
		t.Fatalf("ToggleCardStatus() unexpected error: %v", err)
	}
}

func TestUpdateCardCardNotFound(t *testing.T) {
	repo := newFakeCardCommandRepository(t, &stubConfig{
		queryRowErr: sql.ErrNoRows,
	})

	_, err := repo.UpdateCard(context.Background(), &requests.UpdateCardRequest{
		CardID:       99,
		UserID:       1,
		CardType:     "credit",
		ExpireDate:   time.Date(2028, 12, 1, 0, 0, 0, 0, time.UTC),
		CVV:          "123",
		CardProvider: "visa",
	})

	assertAppErrorCode(t, err, http.StatusNotFound, "card not found")
}

func TestUpdateCardInternalError(t *testing.T) {
	repo := newFakeCardCommandRepository(t, &stubConfig{
		queryRowErr: errors.New("connection reset"),
	})

	_, err := repo.UpdateCard(context.Background(), &requests.UpdateCardRequest{
		CardID:       1,
		UserID:       1,
		CardType:     "credit",
		ExpireDate:   time.Date(2028, 12, 1, 0, 0, 0, 0, time.UTC),
		CVV:          "123",
		CardProvider: "visa",
	})

	assertAppErrorCode(t, err, http.StatusInternalServerError, "Failed to update card")
}

func TestUpdateCardSuccess(t *testing.T) {
	repo := newFakeCardCommandRepository(t, &stubConfig{})

	if _, err := repo.UpdateCard(context.Background(), &requests.UpdateCardRequest{
		CardID:       1,
		UserID:       1,
		CardType:     "credit",
		ExpireDate:   time.Date(2028, 12, 1, 0, 0, 0, 0, time.UTC),
		CVV:          "123",
		CardProvider: "visa",
	}); err != nil {
		t.Fatalf("UpdateCard() unexpected error: %v", err)
	}
}

func TestTrashedCardCardNotFound(t *testing.T) {
	repo := newFakeCardCommandRepository(t, &stubConfig{
		queryRowErr: sql.ErrNoRows,
	})

	_, err := repo.TrashedCard(context.Background(), 99)

	assertAppErrorCode(t, err, http.StatusNotFound, "card not found")
}

func TestTrashedCardInternalError(t *testing.T) {
	repo := newFakeCardCommandRepository(t, &stubConfig{
		queryRowErr: errors.New("connection reset"),
	})

	_, err := repo.TrashedCard(context.Background(), 1)

	assertAppErrorCode(t, err, http.StatusInternalServerError, "Failed to trash card")
}

func TestTrashedCardSuccess(t *testing.T) {
	repo := newFakeCardCommandRepository(t, &stubConfig{})

	if _, err := repo.TrashedCard(context.Background(), 1); err != nil {
		t.Fatalf("TrashedCard() unexpected error: %v", err)
	}
}

func TestRestoreCardCardNotFound(t *testing.T) {
	repo := newFakeCardCommandRepository(t, &stubConfig{
		queryRowErr: sql.ErrNoRows,
	})

	_, err := repo.RestoreCard(context.Background(), 99)

	assertAppErrorCode(t, err, http.StatusNotFound, "card not found")
}

func TestRestoreCardInternalError(t *testing.T) {
	repo := newFakeCardCommandRepository(t, &stubConfig{
		queryRowErr: errors.New("connection reset"),
	})

	_, err := repo.RestoreCard(context.Background(), 1)

	assertAppErrorCode(t, err, http.StatusInternalServerError, "Failed to restore card")
}

func TestRestoreCardSuccess(t *testing.T) {
	repo := newFakeCardCommandRepository(t, &stubConfig{})

	if _, err := repo.RestoreCard(context.Background(), 1); err != nil {
		t.Fatalf("RestoreCard() unexpected error: %v", err)
	}
}

func TestUpdateCreditLimitCardNotFound(t *testing.T) {
	repo := newFakeCardCommandRepository(t, &stubConfig{
		queryRowErr: sql.ErrNoRows,
	})

	_, err := repo.UpdateCreditLimit(context.Background(), &requests.UpdateCreditLimitRequest{
		CardID:      99,
		CreditLimit: 5_000_000,
	})

	assertAppErrorCode(t, err, http.StatusNotFound, "card not found")
}

func TestUpdateCreditLimitInternalError(t *testing.T) {
	repo := newFakeCardCommandRepository(t, &stubConfig{
		queryRowErr: errors.New("connection reset"),
	})

	_, err := repo.UpdateCreditLimit(context.Background(), &requests.UpdateCreditLimitRequest{
		CardID:      1,
		CreditLimit: 5_000_000,
	})

	assertAppErrorCode(t, err, http.StatusInternalServerError, "Failed to update credit limit")
}

func TestUpdateCreditLimitSuccess(t *testing.T) {
	repo := newFakeCardCommandRepository(t, &stubConfig{})

	if _, err := repo.UpdateCreditLimit(context.Background(), &requests.UpdateCreditLimitRequest{
		CardID:      1,
		CreditLimit: 5_000_000,
	}); err != nil {
		t.Fatalf("UpdateCreditLimit() unexpected error: %v", err)
	}
}

func TestRedeemPointsCardNotFound(t *testing.T) {
	// The redeem query returns no rows AND the card itself does not exist,
	// so the fallback GetCardByID also reports ErrNoRows -> 404.
	repo := newFakeCardCommandRepository(t, &stubConfig{
		queryRowErr: sql.ErrNoRows,
		getCardErr:  sql.ErrNoRows,
	})

	_, err := repo.RedeemPoints(context.Background(), &requests.RedeemPointsRequest{
		CardID: 99,
		Points: 100,
	})

	assertAppErrorCode(t, err, http.StatusNotFound, "card not found")
}

func TestRedeemPointsInsufficientRewardPoints(t *testing.T) {
	// The redeem query returns no rows (guard reward_points >= points failed)
	// but the card exists, so the fallback GetCardByID succeeds -> 400.
	repo := newFakeCardCommandRepository(t, &stubConfig{
		queryRowErr: sql.ErrNoRows,
		getCardOK:   true,
	})

	_, err := repo.RedeemPoints(context.Background(), &requests.RedeemPointsRequest{
		CardID: 1,
		Points: 100,
	})

	assertAppErrorCode(t, err, http.StatusBadRequest, "insufficient reward points")
}

func TestRedeemPointsFallbackInternalError(t *testing.T) {
	// The redeem query returns no rows, but the GetCardByID fallback itself
	// fails with a non-ErrNoRows error (e.g. connection reset) -> 500.
	repo := newFakeCardCommandRepository(t, &stubConfig{
		queryRowErr: sql.ErrNoRows,
		getCardErr:  errors.New("connection reset"),
	})

	_, err := repo.RedeemPoints(context.Background(), &requests.RedeemPointsRequest{
		CardID: 1,
		Points: 100,
	})

	assertAppErrorCode(t, err, http.StatusInternalServerError, "Failed to redeem reward points")
}

func TestRedeemPointsInternalError(t *testing.T) {
	repo := newFakeCardCommandRepository(t, &stubConfig{
		queryRowErr: errors.New("connection reset"),
	})

	_, err := repo.RedeemPoints(context.Background(), &requests.RedeemPointsRequest{
		CardID: 1,
		Points: 100,
	})

	assertAppErrorCode(t, err, http.StatusInternalServerError, "Failed to redeem reward points")
}

func TestRedeemPointsSuccess(t *testing.T) {
	repo := newFakeCardCommandRepository(t, &stubConfig{})

	if _, err := repo.RedeemPoints(context.Background(), &requests.RedeemPointsRequest{
		CardID: 1,
		Points: 100,
	}); err != nil {
		t.Fatalf("RedeemPoints() unexpected error: %v", err)
	}
}
