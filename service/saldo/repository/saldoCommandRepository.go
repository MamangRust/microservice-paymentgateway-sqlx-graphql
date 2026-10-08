package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jmoiron/sqlx"

	database "github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database"
	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
)

// SQL source of truth (formerly database/query/saldo.sql).
const (
	createSaldo = `WITH inserted AS (
    INSERT INTO saldos (
        card_number,
        total_balance,
        created_at,
        updated_at
    )
    VALUES (
        $1,
        $2,
        current_timestamp,
        current_timestamp
    )
    ON CONFLICT (card_number) WHERE deleted_at IS NULL
    DO UPDATE SET
        total_balance = EXCLUDED.total_balance,
        updated_at = current_timestamp
    WHERE saldos.total_balance = 0 AND EXCLUDED.total_balance > 0
    RETURNING
        saldo_id,
        card_number,
        total_balance,
        withdraw_amount,
        withdraw_time,
        created_at,
        updated_at
)
SELECT saldo_id, card_number, total_balance, withdraw_amount,
       withdraw_time, created_at, updated_at
FROM inserted
UNION ALL
SELECT saldo_id, card_number, total_balance, withdraw_amount,
       withdraw_time, created_at, updated_at
FROM saldos
WHERE card_number = $1
  AND deleted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM inserted)`

	createSaldoIfNotExists = `INSERT INTO
    saldos (
        card_number,
        total_balance,
        created_at,
        updated_at
    )
VALUES (
        $1,
        $2,
        current_timestamp,
        current_timestamp
    )
ON CONFLICT (card_number) WHERE deleted_at IS NULL
DO NOTHING`

	updateSaldoBalance = `UPDATE saldos
SET
    total_balance = $2,
    updated_at = current_timestamp
WHERE
    card_number = $1
    AND deleted_at IS NULL
RETURNING
    saldo_id,
    card_number,
    total_balance,
    withdraw_amount,
    withdraw_time,
    created_at,
    updated_at`

	updateSaldoWithdrawMetadata = `UPDATE saldos
SET
    withdraw_amount = $2,
    withdraw_time = $3,
    updated_at = current_timestamp
WHERE card_number = $1 AND deleted_at IS NULL
RETURNING saldo_id, card_number, total_balance, withdraw_amount, withdraw_time, created_at, updated_at`

	trashSaldo = `UPDATE saldos
SET
    deleted_at = current_timestamp
WHERE
    saldo_id = $1
    AND deleted_at IS NULL
RETURNING
    saldo_id,
    card_number,
    total_balance,
    withdraw_amount,
    withdraw_time,
    created_at,
    updated_at,
    deleted_at`

	restoreSaldo = `UPDATE saldos
SET
    deleted_at = NULL
WHERE
    saldo_id = $1
    AND deleted_at IS NOT NULL
RETURNING
    saldo_id,
    card_number,
    total_balance,
    withdraw_amount,
    withdraw_time,
    created_at,
    updated_at,
    deleted_at`

	deleteSaldoPermanently = `DELETE FROM saldos WHERE saldo_id = $1 AND deleted_at IS NOT NULL`

	restoreAllSaldos = `UPDATE saldos
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL`

	deleteAllPermanentSaldos = `DELETE FROM saldos WHERE deleted_at IS NOT NULL`

	// Named queries kept for SQL source-of-truth parity (not yet wired to a
	// repository method — UpdateSaldo is routed through the immutable ledger
	// adjustment path instead):
	updateSaldo = `UPDATE saldos
SET
    card_number = $2,
    total_balance = $3,
    updated_at = current_timestamp
WHERE
    saldo_id = $1
    AND deleted_at IS NULL
RETURNING
    saldo_id,
    card_number,
    total_balance,
    withdraw_amount,
    withdraw_time,
    created_at,
    updated_at`

	debitSaldo = `UPDATE saldos
SET
    total_balance = total_balance - $2,
    updated_at = current_timestamp
WHERE
    card_number = $1
    AND deleted_at IS NULL
    AND $2 > 0
    AND total_balance >= $2
RETURNING
    saldo_id,
    card_number,
    total_balance,
    withdraw_amount,
    withdraw_time,
    created_at,
    updated_at`

	creditSaldo = `UPDATE saldos
SET
    total_balance = total_balance + $2,
    updated_at = current_timestamp
WHERE
    card_number = $1
    AND deleted_at IS NULL
    AND $2 > 0
RETURNING
    saldo_id,
    card_number,
    total_balance,
    withdraw_amount,
    withdraw_time,
    created_at,
    updated_at`

	updateSaldoWithdraw = `UPDATE saldos
SET
    withdraw_amount = $2,
    total_balance = total_balance - $2,
    withdraw_time = $3,
    updated_at = current_timestamp
WHERE
    card_number = $1
    AND deleted_at IS NULL
    AND total_balance >= $2
RETURNING
    saldo_id,
    card_number,
    total_balance,
    withdraw_amount,
    withdraw_time,
    created_at,
    updated_at`
)

type saldoCommandRepository struct {
	db *sqlx.DB
}

func NewSaldoCommandRepository(db *sqlx.DB) SaldoCommandRepository {
	return &saldoCommandRepository{
		db: db,
	}
}

func (r *saldoCommandRepository) CreateSaldo(ctx context.Context, request *requests.CreateSaldoRequest) (*db.CreateSaldoRow, error) {
	var row db.CreateSaldoRow
	err := r.db.GetContext(ctx, &row, createSaldo,
		request.CardNumber,
		amountToInt64(request.TotalBalance),
	)

	if err != nil {
		// The SQL is an atomic idempotent upsert guarded by the partial unique
		// index idx_saldos_card_number_active (card_number WHERE deleted_at IS
		// NULL). If a unique violation still surfaces (e.g. constraint mismatch
		// with the index predicate), map it to 409 instead of a generic 500.
		if database.IsUniqueViolation(err) {
			return nil, sharedErrors.NewConflictError("saldo record already exists").WithInternal(err)
		}

		// READ COMMITTED snapshot race: the CTE's INSERT ... ON CONFLICT DO
		// NOTHING blocks on the unique index while the concurrent card-created
		// Kafka consumer commits, then skips (no inserted row), but the
		// UNION-ALL SELECT still uses the statement-start snapshot and cannot
		// see the just-committed row -> no rows in result set. Re-read the row
		// in a fresh statement to return the winner's record.
		if errors.Is(err, sql.ErrNoRows) {
			var existing db.Saldo
			getErr := r.db.GetContext(ctx, &existing, getSaldoByCardNumber, request.CardNumber)
			if getErr != nil {
				return nil, sharedErrors.ErrFailed("create saldo record").WithInternal(getErr)
			}
			return &db.CreateSaldoRow{
				SaldoID:        existing.SaldoID,
				CardNumber:     existing.CardNumber,
				TotalBalance:   existing.TotalBalance,
				WithdrawAmount: existing.WithdrawAmount,
				WithdrawTime:   existing.WithdrawTime,
				CreatedAt:      existing.CreatedAt,
				UpdatedAt:      existing.UpdatedAt,
			}, nil
		}

		return nil, sharedErrors.ErrFailed("create saldo record").WithInternal(err)
	}

	return &row, nil
}

func (r *saldoCommandRepository) CreateSaldoIfNotExists(ctx context.Context, request *requests.CreateSaldoRequest) error {
	_, err := r.db.ExecContext(ctx, createSaldoIfNotExists,
		request.CardNumber,
		amountToInt64(request.TotalBalance),
	)
	return err
}

func (r *saldoCommandRepository) UpdateSaldo(ctx context.Context, request *requests.UpdateSaldoRequest) (*db.UpdateSaldoRow, error) {
	if request.SaldoID == nil {
		return nil, sharedErrors.NewBadRequestError("saldo ID is required")
	}

	var current db.GetSaldoByIDRow
	err := r.db.GetContext(ctx, &current, getSaldoByID, int32(*request.SaldoID))
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "saldo record", "update saldo record")
	}
	// Changing the card association of an active saldo is not allowed; balance
	// changes must go through the immutable ledger adjustment path.
	if current.CardNumber != request.CardNumber {
		return nil, sharedErrors.NewConflictError("cannot change card number of an existing saldo").WithInternal(fmt.Errorf("card mismatch: %s != %s", current.CardNumber, request.CardNumber))
	}

	target := amountToInt64(request.TotalBalance)
	delta := target - current.TotalBalance
	if delta == 0 {
		return &db.UpdateSaldoRow{
			SaldoID: current.SaldoID, CardNumber: current.CardNumber, TotalBalance: current.TotalBalance,
			WithdrawAmount: current.WithdrawAmount, WithdrawTime: current.WithdrawTime,
			CreatedAt: current.CreatedAt, UpdatedAt: current.UpdatedAt,
		}, nil
	}

	// The operation identity includes the target balance so a retry with the
	// same target is a replay no-op, while a legitimate second update with a
	// different target produces a distinct immutable adjustment entry.
	adjusted, err := db.ApplyLedgerAdjustment(ctx, r.db, db.LedgerAdjustmentParams{
		CardNumber:  request.CardNumber,
		Delta:       delta,
		OperationID: fmt.Sprintf("saldo-update:%d:%d", *request.SaldoID, target),
		SourceType:  "saldo_update",
		SourceID:    fmt.Sprintf("%d", *request.SaldoID),
		Note:        "legacy absolute update routed through immutable ledger adjustment",
	})
	if err != nil {
		if errors.Is(err, db.ErrLedgerOperationConflict) {
			return nil, sharedErrors.NewConflictError("saldo update operation ID reused with different details").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("update saldo record").WithInternal(err)
	}

	return &db.UpdateSaldoRow{
		SaldoID: adjusted.SaldoID, CardNumber: adjusted.CardNumber, TotalBalance: adjusted.TotalBalance,
		WithdrawAmount: current.WithdrawAmount, WithdrawTime: current.WithdrawTime,
		CreatedAt: current.CreatedAt, UpdatedAt: current.UpdatedAt,
	}, nil
}

func (r *saldoCommandRepository) UpdateSaldoBalance(ctx context.Context, request *requests.UpdateSaldoBalance) (*db.UpdateSaldoBalanceRow, error) {
	var row db.UpdateSaldoBalanceRow
	err := r.db.GetContext(ctx, &row, updateSaldoBalance,
		request.CardNumber,
		amountToInt64(request.TotalBalance),
	)

	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "saldo record", "update saldo balance")
	}

	return &row, nil
}

func (r *saldoCommandRepository) DebitSaldo(ctx context.Context, request *requests.DebitSaldoRequest) (*db.DebitSaldoRow, error) {
	amount := amountToInt64(request.Amount)

	operationID := request.OperationID
	if operationID == "" {
		operationID = uuid.NewString()
	}
	sourceType := request.SourceType
	if sourceType == "" {
		sourceType = "unknown"
	}
	res, err := db.DebitSaldoWithLedger(ctx, r.db, db.LedgerMutationParams{
		CardNumber:  request.CardNumber,
		Amount:      amount,
		OperationID: operationID,
		SourceType:  sourceType,
		SourceID:    request.SourceID,
	})
	if err != nil {
		if errors.Is(err, db.ErrLedgerOperationConflict) {
			return nil, sharedErrors.NewConflictError("ledger operation ID was reused with different mutation details").WithInternal(err)
		}
		// A no-rows result here means the atomic predicate total_balance >= $2
		// rejected the debit: the card exists (callers resolve it first) but the
		// balance is insufficient. Surface that as a 409 domain error instead of
		// a misleading 404 not-found. Any other error stays a generic failure.
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.NewConflictError("insufficient balance").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("debit saldo").WithInternal(err)
	}
	return res, nil
}

func (r *saldoCommandRepository) CreditSaldo(ctx context.Context, request *requests.CreditSaldoRequest) (*db.CreditSaldoRow, error) {
	amount := amountToInt64(request.Amount)

	operationID := request.OperationID
	if operationID == "" {
		operationID = uuid.NewString()
	}
	sourceType := request.SourceType
	if sourceType == "" {
		sourceType = "unknown"
	}
	res, err := db.CreditSaldoWithLedger(ctx, r.db, db.LedgerMutationParams{
		CardNumber:  request.CardNumber,
		Amount:      amount,
		OperationID: operationID,
		SourceType:  sourceType,
		SourceID:    request.SourceID,
	})
	if err != nil {
		if errors.Is(err, db.ErrLedgerOperationConflict) {
			return nil, sharedErrors.NewConflictError("ledger operation ID was reused with different mutation details").WithInternal(err)
		}
		// No-rows here means the card has no active saldo record (amount is
		// validated > 0 upstream), so keep the not-found mapping.
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "saldo record", "credit saldo")
	}
	return res, nil
}

func (r *saldoCommandRepository) ApplySaldoAdjustment(ctx context.Context, request *requests.ApplySaldoAdjustmentRequest) (*db.SaldoAdjustmentRow, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}
	res, err := db.ApplyLedgerAdjustment(ctx, r.db, db.LedgerAdjustmentParams{
		CardNumber:  request.CardNumber,
		Delta:       request.Delta,
		OperationID: request.OperationID,
		SourceType:  request.SourceType,
		SourceID:    request.SourceID,
		Note:        request.Note,
	})
	if err != nil {
		if errors.Is(err, db.ErrLedgerOperationConflict) {
			return nil, sharedErrors.NewConflictError("adjustment operation ID was reused with different mutation details").WithInternal(err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.NewConflictError("saldo adjustment rejected").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("apply saldo adjustment").WithInternal(err)
	}
	return res, nil
}

func (r *saldoCommandRepository) ResolveReconciliation(ctx context.Context, queueID int64, operationID, note string) error {
	if queueID <= 0 || operationID == "" {
		return sharedErrors.NewBadRequestError("queue ID and operation ID are required")
	}
	if err := db.ResolveReconciliation(ctx, r.db, queueID, operationID, note); err != nil {
		return sharedErrors.ErrNoRowsOrFailed(err, "reconciliation queue item", "resolve reconciliation")
	}
	return nil
}

func (r *saldoCommandRepository) UpdateSaldoWithdraw(ctx context.Context, request *requests.UpdateSaldoWithdraw) (*db.UpdateSaldoWithdrawRow, error) {
	withdrawAmount := optionalAmountToInt64(request.WithdrawAmount)

	if withdrawAmount == nil || request.WithdrawTime == nil {
		return nil, sharedErrors.NewBadRequestError("withdraw amount and time are required for ledger-backed withdrawal")
	}

	var withdrawTime pgtype.Timestamp
	withdrawTime = pgtype.Timestamp{
		Time:  *request.WithdrawTime,
		Valid: true,
	}

	operationID := "withdraw:" + request.CardNumber + ":" + request.WithdrawTime.UTC().Format("20060102150405.000000000") + ":" + fmt.Sprint(*withdrawAmount)
	debit, err := r.DebitSaldo(ctx, &requests.DebitSaldoRequest{
		CardNumber:  request.CardNumber,
		Amount:      int(*withdrawAmount),
		OperationID: operationID,
		SourceType:  "withdraw",
		SourceID:    request.CardNumber,
	})
	if err != nil {
		return nil, err
	}
	var res db.UpdateSaldoWithdrawMetadataRow
	err = r.db.GetContext(ctx, &res, updateSaldoWithdrawMetadata,
		request.CardNumber,
		withdrawAmount,
		withdrawTime,
	)
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "saldo record", "update saldo withdrawal metadata")
	}
	return &db.UpdateSaldoWithdrawRow{
		SaldoID: debit.SaldoID, CardNumber: debit.CardNumber, TotalBalance: debit.TotalBalance,
		WithdrawAmount: res.WithdrawAmount, WithdrawTime: res.WithdrawTime,
		CreatedAt: res.CreatedAt, UpdatedAt: res.UpdatedAt,
	}, nil
}

func (r *saldoCommandRepository) TrashedSaldo(ctx context.Context, saldo_id int) (*db.Saldo, error) {
	var row db.Saldo
	err := r.db.GetContext(ctx, &row, trashSaldo, int32(saldo_id))
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "saldo record", "trash saldo record")
	}
	return &row, nil
}

func (r *saldoCommandRepository) RestoreSaldo(ctx context.Context, saldo_id int) (*db.Saldo, error) {
	var row db.Saldo
	err := r.db.GetContext(ctx, &row, restoreSaldo, int32(saldo_id))
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "saldo record", "restore saldo record")
	}
	return &row, nil
}

func (r *saldoCommandRepository) DeleteSaldoPermanent(ctx context.Context, saldo_id int) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteSaldoPermanently, int32(saldo_id)); err != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(err, "saldo record", "delete saldo record permanently")
	}
	return true, nil
}

func (r *saldoCommandRepository) RestoreAllSaldo(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, restoreAllSaldos); err != nil {
		return false, sharedErrors.ErrFailed("restore all saldo records").WithInternal(err)
	}

	return true, nil
}

func (r *saldoCommandRepository) DeleteAllSaldoPermanent(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteAllPermanentSaldos); err != nil {
		return false, sharedErrors.ErrFailed("delete all saldo records permanently").WithInternal(err)
	}

	return true, nil
}
