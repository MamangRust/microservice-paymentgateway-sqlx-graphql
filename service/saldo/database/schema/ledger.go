package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jmoiron/sqlx"
)

// pgTimestampValue is an alias so NullTime conversion returns the pgtype used
// across the module's row structs.
type pgTimestampValue = pgtype.Timestamp

// ErrLedgerOperationConflict indicates that an operation ID was reused with
// different mutation details. Replaying such a request must never be treated as
// a successful no-op.
var ErrLedgerOperationConflict = errors.New("ledger operation identity conflict")

// LedgerMutationParams carries the durable identity for a saldo mutation.
type LedgerMutationParams struct {
	CardNumber  string
	Amount      int64
	OperationID string
	SourceType  string
	SourceID    string
}

// LedgerAdjustmentParams represents an append-only signed correction. Positive
// delta credits the saldo; negative delta debits it.
type LedgerAdjustmentParams struct {
	CardNumber  string
	Delta       int64
	OperationID string
	SourceType  string
	SourceID    string
	Note        string
}

// LedgerEntry is the immutable audit representation of a balance change.
type LedgerEntry struct {
	EntryID       int64   `db:"entry_id"`
	OperationID   string  `db:"operation_id"`
	CardNumber    string  `db:"card_number"`
	Direction     string  `db:"direction"`
	Amount        int64   `db:"amount"`
	Delta         int64   `db:"delta"`
	BalanceBefore int64   `db:"balance_before"`
	BalanceAfter  int64   `db:"balance_after"`
	SourceType    string  `db:"source_type"`
	SourceID      *string `db:"source_id"`
	Note          *string `db:"note"`
	CreatedAt     time.Time
}

// ReconciliationRow compares current saldo with opening + ledger deltas.
type ReconciliationRow struct {
	SaldoID        int32  `db:"saldo_id"`
	CardNumber     string `db:"card_number"`
	CurrentBalance int64  `db:"current_balance"`
	LedgerBalance  int64  `db:"ledger_balance"`
	Difference     int64  `db:"difference"`
	LedgerEntries  int64  `db:"ledger_entries"`
}

// ReconciliationQueueRow is the durable operator-facing mismatch record.
type ReconciliationQueueRow struct {
	QueueID               int64      `db:"queue_id"`
	SaldoID               int32      `db:"saldo_id"`
	CardNumber            string     `db:"card_number"`
	CurrentBalance        int64      `db:"current_balance"`
	LedgerBalance         int64      `db:"ledger_balance"`
	Difference            int64      `db:"difference"`
	LedgerEntries         int64      `db:"ledger_entries"`
	Status                string     `db:"status"`
	MismatchCount         int64      `db:"mismatch_count"`
	FirstSeenAt           time.Time  `db:"first_seen_at"`
	LastSeenAt            time.Time  `db:"last_seen_at"`
	ResolvedAt            *time.Time `db:"resolved_at"`
	ResolutionOperationID *string    `db:"resolution_operation_id"`
	ResolutionNote        *string    `db:"resolution_note"`
}

// SaldoAdjustmentRow is returned by an append-only correction.
type SaldoAdjustmentRow struct {
	SaldoID      int32  `db:"saldo_id"`
	CardNumber   string `db:"card_number"`
	TotalBalance int64  `db:"total_balance"`
}

// balanceRow mirrors the RETURNING clause of the ledger mutation queries so
// both mutations and the replay lookups share one scan type.
type balanceRow struct {
	SaldoID        int32        `db:"saldo_id"`
	CardNumber     string       `db:"card_number"`
	TotalBalance   int64        `db:"total_balance"`
	WithdrawAmount *int64       `db:"withdraw_amount"`
	WithdrawTime   sql.NullTime `db:"withdraw_time"`
	CreatedAt      sql.NullTime `db:"created_at"`
	UpdatedAt      sql.NullTime `db:"updated_at"`
}

func (r balanceRow) toDebitSaldoRow() *DebitSaldoRow {
	return &DebitSaldoRow{
		SaldoID: r.SaldoID, CardNumber: r.CardNumber, TotalBalance: r.TotalBalance,
		WithdrawAmount: r.WithdrawAmount, WithdrawTime: pgTimestamp(r.WithdrawTime),
		CreatedAt: pgTimestamp(r.CreatedAt), UpdatedAt: pgTimestamp(r.UpdatedAt),
	}
}

func (r balanceRow) toCreditSaldoRow() *CreditSaldoRow {
	return &CreditSaldoRow{
		SaldoID: r.SaldoID, CardNumber: r.CardNumber, TotalBalance: r.TotalBalance,
		WithdrawAmount: r.WithdrawAmount, WithdrawTime: pgTimestamp(r.WithdrawTime),
		CreatedAt: pgTimestamp(r.CreatedAt), UpdatedAt: pgTimestamp(r.UpdatedAt),
	}
}

func (r balanceRow) toUpdateSaldoWithdrawRow(from *UpdateSaldoWithdrawMetadataRow) *UpdateSaldoWithdrawRow {
	return &UpdateSaldoWithdrawRow{
		SaldoID: r.SaldoID, CardNumber: r.CardNumber, TotalBalance: r.TotalBalance,
		WithdrawAmount: from.WithdrawAmount, WithdrawTime: from.WithdrawTime,
		CreatedAt: from.CreatedAt, UpdatedAt: from.UpdatedAt,
	}
}

func pgTimestamp(t sql.NullTime) pgtype.Timestamp {
	if !t.Valid {
		return pgtype.Timestamp{}
	}
	return pgtype.Timestamp{Time: t.Time, Valid: true}
}

// GetSaldoByCardNumber returns the active saldo record for a card. Used by the
// ledger replay paths after a no-rows mutation result.
func GetSaldoByCardNumber(ctx context.Context, q sqlx.QueryerContext, cardNumber string) (*Saldo, error) {
	const query = `SELECT * FROM saldos
WHERE card_number = $1 AND deleted_at IS NULL
ORDER BY saldo_id DESC
LIMIT 1`
	var saldo Saldo
	if err := sqlx.GetContext(ctx, q, &saldo, query, cardNumber); err != nil {
		return nil, err
	}
	return &saldo, nil
}

// existingLedgerMutation reports whether an identical prior mutation exists for
// the operation identity (idempotent replay).
func existingLedgerMutation(ctx context.Context, q sqlx.QueryerContext, arg LedgerMutationParams, direction string) (bool, error) {
	var row struct {
		Amount     int64   `db:"amount"`
		SourceType string  `db:"source_type"`
		SourceID   *string `db:"source_id"`
	}
	err := sqlx.GetContext(ctx, q, &row, `
		SELECT amount, source_type, source_id
		FROM balance_ledger
		WHERE operation_id = $1 AND card_number = $2 AND direction = $3
	`, arg.OperationID, arg.CardNumber, direction)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	amount := row.Amount
	sourceType := row.SourceType
	sourceID := row.SourceID
	requestedSourceID := arg.SourceID
	storedSourceID := ""
	if sourceID != nil {
		storedSourceID = *sourceID
	}
	if amount != arg.Amount || sourceType != arg.SourceType || storedSourceID != requestedSourceID {
		return false, fmt.Errorf("%w: operation_id=%s", ErrLedgerOperationConflict, arg.OperationID)
	}
	return true, nil
}

func ledgerWithdrawAmount(value *int64) *int64 {
	if value == nil {
		return nil
	}
	converted := *value
	return &converted
}

func DebitSaldoWithLedger(ctx context.Context, db *sqlx.DB, arg LedgerMutationParams) (*DebitSaldoRow, error) {
	return debitSaldoWithLedger(ctx, db, arg)
}

func debitSaldoWithLedger(ctx context.Context, q sqlx.QueryerContext, arg LedgerMutationParams) (*DebitSaldoRow, error) {
	const query = `
		WITH candidate AS MATERIALIZED (
			SELECT saldo_id, card_number, total_balance, withdraw_amount,
			       withdraw_time, created_at, updated_at
			FROM saldos
			WHERE card_number = $1 AND deleted_at IS NULL
			FOR UPDATE
		), inserted AS (
			INSERT INTO balance_ledger (
				operation_id, card_number, direction, amount, delta,
				balance_before, balance_after, source_type, source_id, note
			)
			SELECT $3, card_number, 'debit', $2::bigint, -$2::bigint, total_balance,
			       total_balance - $2::bigint, $4, NULLIF($5, ''), NULL
			FROM candidate
			WHERE $2::bigint > 0 AND total_balance >= $2::bigint
			ON CONFLICT (operation_id, card_number, direction) DO NOTHING
			RETURNING card_number
		)
		UPDATE saldos s
		SET total_balance = s.total_balance - $2::bigint,
		    updated_at = current_timestamp
		FROM inserted i
		WHERE s.card_number = i.card_number AND s.deleted_at IS NULL
		RETURNING s.saldo_id, s.card_number, s.total_balance,
		          s.withdraw_amount, s.withdraw_time, s.created_at, s.updated_at
	`
	var result balanceRow
	err := sqlx.GetContext(ctx, q, &result, query, arg.CardNumber, arg.Amount, arg.OperationID, arg.SourceType, arg.SourceID)
	if errors.Is(err, sql.ErrNoRows) {
		if existing, lookupErr := existingLedgerMutation(ctx, q, arg, "debit"); lookupErr != nil {
			return nil, lookupErr
		} else if existing {
			current, currentErr := GetSaldoByCardNumber(ctx, q, arg.CardNumber)
			if currentErr != nil {
				return nil, currentErr
			}
			return &DebitSaldoRow{SaldoID: current.SaldoID, CardNumber: current.CardNumber, TotalBalance: current.TotalBalance, WithdrawAmount: ledgerWithdrawAmount(current.WithdrawAmount), WithdrawTime: current.WithdrawTime, CreatedAt: current.CreatedAt, UpdatedAt: current.UpdatedAt}, nil
		}
	}
	if err != nil {
		return nil, err
	}
	return result.toDebitSaldoRow(), nil
}

// CreditSaldoWithLedger atomically updates saldo and appends one credit ledger entry.
func CreditSaldoWithLedger(ctx context.Context, db *sqlx.DB, arg LedgerMutationParams) (*CreditSaldoRow, error) {
	return creditSaldoWithLedger(ctx, db, arg)
}

func creditSaldoWithLedger(ctx context.Context, q sqlx.QueryerContext, arg LedgerMutationParams) (*CreditSaldoRow, error) {
	const query = `
		WITH candidate AS MATERIALIZED (
			SELECT saldo_id, card_number, total_balance, withdraw_amount,
			       withdraw_time, created_at, updated_at
			FROM saldos
			WHERE card_number = $1 AND deleted_at IS NULL
			FOR UPDATE
		), inserted AS (
			INSERT INTO balance_ledger (
				operation_id, card_number, direction, amount, delta,
				balance_before, balance_after, source_type, source_id, note
			)
			SELECT $3, card_number, 'credit', $2::bigint, $2::bigint, total_balance,
			       total_balance + $2::bigint, $4, NULLIF($5, ''), NULL
			FROM candidate
			WHERE $2::bigint > 0
			ON CONFLICT (operation_id, card_number, direction) DO NOTHING
			RETURNING card_number
		)
		UPDATE saldos s
		SET total_balance = s.total_balance + $2::bigint,
		    updated_at = current_timestamp
		FROM inserted i
		WHERE s.card_number = i.card_number AND s.deleted_at IS NULL
		RETURNING s.saldo_id, s.card_number, s.total_balance,
		          s.withdraw_amount, s.withdraw_time, s.created_at, s.updated_at
	`
	var result balanceRow
	err := sqlx.GetContext(ctx, q, &result, query, arg.CardNumber, arg.Amount, arg.OperationID, arg.SourceType, arg.SourceID)
	if errors.Is(err, sql.ErrNoRows) {
		if existing, lookupErr := existingLedgerMutation(ctx, q, arg, "credit"); lookupErr != nil {
			return nil, lookupErr
		} else if existing {
			current, currentErr := GetSaldoByCardNumber(ctx, q, arg.CardNumber)
			if currentErr != nil {
				return nil, currentErr
			}
			return &CreditSaldoRow{SaldoID: current.SaldoID, CardNumber: current.CardNumber, TotalBalance: current.TotalBalance, WithdrawAmount: ledgerWithdrawAmount(current.WithdrawAmount), WithdrawTime: current.WithdrawTime, CreatedAt: current.CreatedAt, UpdatedAt: current.UpdatedAt}, nil
		}
	}
	if err != nil {
		return nil, err
	}
	return result.toCreditSaldoRow(), nil
}

func existingLedgerAdjustment(ctx context.Context, q sqlx.QueryerContext, arg LedgerAdjustmentParams) (bool, error) {
	var row struct {
		Amount     int64   `db:"amount"`
		Delta      int64   `db:"delta"`
		SourceType string  `db:"source_type"`
		SourceID   *string `db:"source_id"`
		Note       *string `db:"note"`
	}
	err := sqlx.GetContext(ctx, q, &row, `
		SELECT amount, delta, source_type, source_id, note
		FROM balance_ledger
		WHERE operation_id = $1 AND card_number = $2 AND direction = 'reversal'
	`, arg.OperationID, arg.CardNumber)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	amount := row.Amount
	delta := row.Delta
	sourceType := row.SourceType
	sourceID := row.SourceID
	note := row.Note
	storedSourceID := ""
	if sourceID != nil {
		storedSourceID = *sourceID
	}
	storedNote := ""
	if note != nil {
		storedNote = *note
	}
	if delta != arg.Delta || amount != absInt64(arg.Delta) || sourceType != arg.SourceType || storedSourceID != arg.SourceID || storedNote != arg.Note {
		return false, fmt.Errorf("%w: operation_id=%s", ErrLedgerOperationConflict, arg.OperationID)
	}
	return true, nil
}

func absInt64(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}

// ApplyLedgerAdjustment changes the current balance only through a new
// immutable reversal entry. delta > 0 credits; delta < 0 debits.
func ApplyLedgerAdjustment(ctx context.Context, db *sqlx.DB, arg LedgerAdjustmentParams) (*SaldoAdjustmentRow, error) {
	if arg.Delta == 0 {
		return nil, errors.New("ledger adjustment delta must not be zero")
	}
	const query = `
		WITH candidate AS MATERIALIZED (
			SELECT saldo_id, card_number, total_balance
			FROM saldos
			WHERE card_number = $1 AND deleted_at IS NULL
			FOR UPDATE
		), inserted AS (
			INSERT INTO balance_ledger (
				operation_id, card_number, direction, amount, delta,
				balance_before, balance_after, source_type, source_id, note
			)
			SELECT $3, card_number, 'reversal', ABS($2::bigint), $2::bigint,
			       total_balance, total_balance + $2::bigint, $4, NULLIF($5, ''), NULLIF($6, '')
			FROM candidate
			WHERE total_balance + $2::bigint >= 0
			ON CONFLICT (operation_id, card_number, direction) DO NOTHING
			RETURNING card_number
		)
		UPDATE saldos s
		SET total_balance = s.total_balance + $2::bigint,
		    updated_at = current_timestamp
		FROM inserted i
		WHERE s.card_number = i.card_number AND s.deleted_at IS NULL
		RETURNING s.saldo_id, s.card_number, s.total_balance
	`
	var result SaldoAdjustmentRow
	err := db.GetContext(ctx, &result, query, arg.CardNumber, arg.Delta, arg.OperationID, arg.SourceType, arg.SourceID, arg.Note)
	if errors.Is(err, sql.ErrNoRows) {
		if existing, lookupErr := existingLedgerAdjustment(ctx, db, arg); lookupErr != nil {
			return nil, lookupErr
		} else if existing {
			current, currentErr := GetSaldoByCardNumber(ctx, db, arg.CardNumber)
			if currentErr != nil {
				return nil, currentErr
			}
			return &SaldoAdjustmentRow{SaldoID: current.SaldoID, CardNumber: current.CardNumber, TotalBalance: current.TotalBalance}, nil
		}
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListReconciliationMismatches returns only accounts whose current balance
// differs from the immutable ledger total.
func ListReconciliationMismatches(ctx context.Context, db *sqlx.DB, limit int32) ([]*ReconciliationRow, error) {
	const query = `
		WITH ledger_totals AS (
			SELECT card_number, COALESCE(SUM(delta), 0) AS ledger_balance,
			       COUNT(*) AS ledger_entries
			FROM balance_ledger
			GROUP BY card_number
		)
		SELECT s.saldo_id, s.card_number, s.total_balance,
		       COALESCE(l.ledger_balance, 0),
		       s.total_balance - COALESCE(l.ledger_balance, 0),
		       COALESCE(l.ledger_entries, 0)
		FROM saldos s
		LEFT JOIN ledger_totals l ON l.card_number = s.card_number
		WHERE s.deleted_at IS NULL
		  AND s.total_balance <> COALESCE(l.ledger_balance, 0)
		ORDER BY ABS(s.total_balance - COALESCE(l.ledger_balance, 0)) DESC, s.saldo_id
		LIMIT $1
	`
	var result []*ReconciliationRow
	if err := db.SelectContext(ctx, &result, query, limit); err != nil {
		return nil, err
	}
	return result, nil
}

func EnqueueReconciliationMismatch(ctx context.Context, db *sqlx.DB, item *ReconciliationRow) error {
	const query = `
		INSERT INTO reconciliation_queue (
			saldo_id, card_number, current_balance, ledger_balance,
			difference, ledger_entries
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (saldo_id) WHERE status IN ('pending', 'investigating')
		DO UPDATE SET current_balance = EXCLUDED.current_balance,
		              ledger_balance = EXCLUDED.ledger_balance,
		              difference = EXCLUDED.difference,
		              ledger_entries = EXCLUDED.ledger_entries,
		              mismatch_count = reconciliation_queue.mismatch_count + 1,
		              last_seen_at = current_timestamp
	`
	_, err := db.ExecContext(ctx, query, item.SaldoID, item.CardNumber, item.CurrentBalance, item.LedgerBalance, item.Difference, item.LedgerEntries)
	return err
}

func ListReconciliationQueue(ctx context.Context, db *sqlx.DB, status string, limit int32) ([]*ReconciliationQueueRow, error) {
	const query = `
		SELECT queue_id, saldo_id, card_number, current_balance, ledger_balance,
		       difference, ledger_entries, status, mismatch_count, first_seen_at,
		       last_seen_at, resolved_at, resolution_operation_id, resolution_note
		FROM reconciliation_queue
		WHERE ($1 = '' OR status = $1)
		ORDER BY CASE status WHEN 'pending' THEN 0 WHEN 'investigating' THEN 1 ELSE 2 END,
		         last_seen_at DESC, queue_id
		LIMIT $2
	`
	var result []*ReconciliationQueueRow
	if err := db.SelectContext(ctx, &result, query, status, limit); err != nil {
		return nil, err
	}
	return result, nil
}

func ResolveReconciliation(ctx context.Context, db *sqlx.DB, queueID int64, operationID, note string) error {
	const query = `
		UPDATE reconciliation_queue q
		SET status = 'resolved', resolved_at = current_timestamp,
		    resolution_operation_id = $2, resolution_note = NULLIF($3, '')
		WHERE q.queue_id = $1
		  AND q.status IN ('pending', 'investigating')
		  AND EXISTS (
			SELECT 1 FROM saldos s
			WHERE s.saldo_id = q.saldo_id AND s.deleted_at IS NULL
			  AND s.total_balance = (
				SELECT COALESCE(SUM(delta), 0) FROM balance_ledger
				WHERE card_number = q.card_number
			  )
		  )
	`
	result, err := db.ExecContext(ctx, query, queueID, operationID, note)
	if err != nil {
		return err
	}
	rowsAffected, rowsErr := result.RowsAffected()
	if rowsErr != nil {
		return rowsErr
	}
	if rowsAffected == 0 {
		return errLedgerNoRows
	}
	return nil
}

// ListLedgerEntries returns immutable audit entries for one card.
func ListLedgerEntries(ctx context.Context, db *sqlx.DB, cardNumber string, limit int32) ([]*LedgerEntry, error) {
	const query = `
		SELECT entry_id, operation_id, card_number, direction, amount, delta,
		       balance_before, balance_after, source_type, source_id, note, created_at
		FROM balance_ledger WHERE card_number = $1
		ORDER BY created_at, entry_id LIMIT $2
	`
	var result []*LedgerEntry
	if err := db.SelectContext(ctx, &result, query, cardNumber, limit); err != nil {
		return nil, err
	}
	return result, nil
}

var errLedgerNoRows = errors.New("ledger target not found")
