package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"

	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/transaction/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/state"
	"github.com/jmoiron/sqlx"
)

const createTransaction = `INSERT INTO
    transactions (
        card_number,
        amount,
        payment_method,
        merchant_id,
        transaction_time,
        created_at,
        updated_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        current_timestamp,
        current_timestamp
    )
RETURNING
    transaction_id,
    transaction_no,
    card_number,
    amount,
    payment_method,
    merchant_id,
    transaction_time,
    status,
    created_at,
    updated_at
`

const updateTransaction = `UPDATE transactions
SET
    card_number = $2,
    amount = $3,
    payment_method = $4,
    merchant_id = $5,
    transaction_time = $6,
    updated_at = current_timestamp
WHERE
    transaction_id = $1
    AND deleted_at IS NULL
RETURNING
    transaction_id,
    transaction_no,
    card_number,
    amount,
    payment_method,
    merchant_id,
    transaction_time,
    status,
    created_at,
    updated_at
`

const updateTransactionStatus = `UPDATE transactions
SET
    status = $2,
    updated_at = current_timestamp
WHERE
    transaction_id = $1
    AND deleted_at IS NULL
RETURNING
    transaction_id,
    transaction_no,
    card_number,
    amount,
    payment_method,
    merchant_id,
    transaction_time,
    status,
    created_at,
    updated_at
`

const trashTransaction = `UPDATE transactions
SET
    deleted_at = current_timestamp
WHERE
    transaction_id = $1
    AND deleted_at IS NULL
RETURNING transaction_id, transaction_no, card_number, amount, payment_method, merchant_id, transaction_time, status, created_at, updated_at, deleted_at, operation_id, failure_reason
`

const restoreTransaction = `UPDATE transactions
SET
    deleted_at = NULL
WHERE
    transaction_id = $1
    AND deleted_at IS NOT NULL
RETURNING transaction_id, transaction_no, card_number, amount, payment_method, merchant_id, transaction_time, status, created_at, updated_at, deleted_at, operation_id, failure_reason
`

const deleteTransactionPermanently = `DELETE FROM transactions
WHERE
    transaction_id = $1
    AND deleted_at IS NOT NULL
`

const restoreAllTransactions = `UPDATE transactions
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL
`

const deleteAllPermanentTransactions = `DELETE FROM transactions WHERE deleted_at IS NOT NULL`

type transactionCommandRepository struct {
	db *sqlx.DB
}

func NewTransactionCommandRepository(db *sqlx.DB) TransactionCommandRepository {
	return &transactionCommandRepository{
		db: db,
	}
}

func (r *transactionCommandRepository) CreateTransaction(ctx context.Context, request *requests.CreateTransactionRequest) (*db.CreateTransactionRow, error) {
	var res db.CreateTransactionRow

	err := sqlx.GetContext(ctx, r.db, &res, createTransaction,
		request.CardNumber,
		int64(request.Amount),
		request.PaymentMethod,
		int32(*request.MerchantID),
		request.TransactionTime,
	)

	if err != nil {
		return nil, sharedErrors.ErrFailed("create transaction").WithInternal(err)
	}

	return &res, nil
}

func (r *transactionCommandRepository) UpdateTransaction(ctx context.Context, request *requests.UpdateTransactionRequest) (*db.UpdateTransactionRow, error) {
	var res db.UpdateTransactionRow

	err := sqlx.GetContext(ctx, r.db, &res, updateTransaction,
		int32(*request.TransactionID),
		request.CardNumber,
		int64(request.Amount),
		request.PaymentMethod,
		int32(*request.MerchantID),
		request.TransactionTime,
	)

	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transaction", "update transaction")
	}

	return &res, nil
}

func (r *transactionCommandRepository) UpdateTransactionStatus(ctx context.Context, request *requests.UpdateTransactionStatus) (*db.UpdateTransactionStatusRow, error) {
	var res db.UpdateTransactionStatusRow

	err := sqlx.GetContext(ctx, r.db, &res, updateTransactionStatus,
		int32(request.TransactionID),
		request.Status,
	)

	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transaction", "update transaction status")
	}

	return &res, nil
}

// GuardStatus atomically transitions a transaction from an expected status to a
// new one. It reports false (with nil error) when the row is not in the expected
// status, i.e. another actor already moved it — callers use this to make
// compensation exactly-once under concurrency.
func (r *transactionCommandRepository) GuardStatus(ctx context.Context, id int, fromStatus, toStatus, reason string) (bool, error) {
	if err := state.CheckTransition(fromStatus, toStatus); err != nil {
		return false, err
	}
	_, err := db.GuardTransactionStatus(ctx, r.db, int32(id), fromStatus, toStatus, reason)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ListStuck returns transactions stuck in a recoverable state for the recovery worker.
func (r *transactionCommandRepository) TransitionStatus(ctx context.Context, id int, fromStatus, toStatus, reason string) (*db.UpdateTransactionStatusRow, error) {
	if err := state.CheckTransition(fromStatus, toStatus); err != nil {
		return nil, err
	}
	return db.GuardTransactionStatus(ctx, r.db, int32(id), fromStatus, toStatus, reason)
}

func (r *transactionCommandRepository) ListStuck(ctx context.Context, olderThan time.Duration, maxRows int32) ([]*db.StuckTransaction, error) {
	return db.ListStuckTransactions(ctx, r.db, olderThan, maxRows)
}

func (r *transactionCommandRepository) TrashedTransaction(ctx context.Context, transaction_id int) (*db.Transaction, error) {
	var res db.Transaction

	err := sqlx.GetContext(ctx, r.db, &res, trashTransaction, int32(transaction_id))
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transaction", "trash transaction")
	}

	return &res, nil
}

func (r *transactionCommandRepository) RestoreTransaction(ctx context.Context, transaction_id int) (*db.Transaction, error) {
	var res db.Transaction

	err := sqlx.GetContext(ctx, r.db, &res, restoreTransaction, int32(transaction_id))
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transaction", "restore transaction")
	}

	return &res, nil
}

func (r *transactionCommandRepository) DeleteTransactionPermanent(ctx context.Context, transaction_id int) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteTransactionPermanently, int32(transaction_id)); err != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(err, "transaction", "delete transaction permanently")
	}

	return true, nil
}

func (r *transactionCommandRepository) RestoreAllTransaction(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, restoreAllTransactions); err != nil {
		return false, sharedErrors.ErrFailed("restore all transactions").WithInternal(err)
	}

	return true, nil
}

func (r *transactionCommandRepository) DeleteAllTransactionPermanent(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteAllPermanentTransactions); err != nil {
		return false, sharedErrors.ErrFailed("delete all transactions permanently").WithInternal(err)
	}

	return true, nil
}
