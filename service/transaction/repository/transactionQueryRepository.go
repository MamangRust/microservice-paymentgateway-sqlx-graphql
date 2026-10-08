package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/transaction/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/jmoiron/sqlx"
)

const getTransactions = `SELECT
    transaction_id,
    transaction_no,
    card_number,
    amount,
    payment_method,
    merchant_id,
    transaction_time,
    status,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM transactions
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR card_number ILIKE '%' || $1 || '%'
        OR payment_method ILIKE '%' || $1 || '%'
        OR status ILIKE '%' || $1 || '%'
    )
ORDER BY transaction_time DESC
LIMIT $2
OFFSET
    $3`

const getActiveTransactions = `SELECT
    transaction_id,
    transaction_no,
    card_number,
    amount,
    payment_method,
    merchant_id,
    transaction_time,
    status,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM transactions
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR card_number ILIKE '%' || $1 || '%'
        OR payment_method ILIKE '%' || $1 || '%'
    )
ORDER BY transaction_time DESC
LIMIT $2
OFFSET
    $3`

const getTrashedTransactions = `SELECT
    transaction_id,
    transaction_no,
    card_number,
    amount,
    payment_method,
    merchant_id,
    transaction_time,
    status,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM transactions
WHERE
    deleted_at IS NOT NULL
    AND (
        $1::TEXT IS NULL
        OR card_number ILIKE '%' || $1 || '%'
        OR payment_method ILIKE '%' || $1 || '%'
    )
ORDER BY transaction_time DESC
LIMIT $2
OFFSET
    $3`

const getTransactionByID = `SELECT
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
FROM transactions
WHERE
    transaction_id = $1
    AND deleted_at IS NULL`

const getTransactionsByCardNumber = `SELECT
    transaction_id,
    transaction_no,
    card_number,
    amount,
    payment_method,
    merchant_id,
    transaction_time,
    status,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM transactions
WHERE
    deleted_at IS NULL
    AND card_number = $1
    AND (
        $2::TEXT IS NULL
        OR payment_method ILIKE '%' || $2 || '%'
        OR status ILIKE '%' || $2 || '%'
    )
ORDER BY transaction_time DESC
LIMIT $3
OFFSET
    $4`

const getTransactionsByMerchantID = `SELECT
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
FROM transactions
WHERE
    merchant_id = $1
    AND deleted_at IS NULL
ORDER BY transaction_time DESC`

const getTransactionsByMerchantIDPaginated = `SELECT
    transaction_id,
    transaction_no,
    card_number,
    amount,
    payment_method,
    merchant_id,
    transaction_time,
    status,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM transactions
WHERE
    deleted_at IS NULL
    AND merchant_id = $1
    AND (
        $2::TEXT IS NULL
        OR card_number ILIKE '%' || $2 || '%'
        OR payment_method ILIKE '%' || $2 || '%'
        OR status ILIKE '%' || $2 || '%'
    )
ORDER BY transaction_time DESC
LIMIT $3
OFFSET
    $4`

type transactionQueryRepository struct {
	db *sqlx.DB
}

func NewTransactionQueryRepository(db *sqlx.DB) TransactionQueryRepository {
	return &transactionQueryRepository{
		db: db,
	}
}

func (r *transactionQueryRepository) FindAllTransactions(ctx context.Context, req *requests.FindAllTransactions) ([]*db.GetTransactionsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var transactions []*db.GetTransactionsRow

	err := r.db.SelectContext(ctx, &transactions, getTransactions,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, sharedErrors.ErrFailed("find all transactions").WithInternal(err)
	}

	return transactions, nil
}

func (r *transactionQueryRepository) FindAllTransactionByCardNumber(ctx context.Context, req *requests.FindAllTransactionCardNumber) ([]*db.GetTransactionsByCardNumberRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var transactions []*db.GetTransactionsByCardNumberRow

	err := r.db.SelectContext(ctx, &transactions, getTransactionsByCardNumber,
		req.CardNumber,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, sharedErrors.ErrFailed("find transactions by card number").WithInternal(err)
	}

	return transactions, nil
}

func (r *transactionQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllTransactions) ([]*db.GetActiveTransactionsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetActiveTransactionsRow

	err := r.db.SelectContext(ctx, &res, getActiveTransactions,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, sharedErrors.ErrFailed("find active transactions").WithInternal(err)
	}

	return res, nil
}

func (r *transactionQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllTransactions) ([]*db.GetTrashedTransactionsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetTrashedTransactionsRow

	err := r.db.SelectContext(ctx, &res, getTrashedTransactions,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, sharedErrors.ErrFailed("find trashed transactions").WithInternal(err)
	}

	return res, nil
}

func (r *transactionQueryRepository) FindById(ctx context.Context, transaction_id int) (*db.GetTransactionByIDRow, error) {
	var res db.GetTransactionByIDRow

	err := r.db.GetContext(ctx, &res, getTransactionByID, int32(transaction_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFoundResponse("transaction").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &res, nil
}

func (r *transactionQueryRepository) FindTransactionByMerchantId(ctx context.Context, merchant_id int) ([]*db.GetTransactionsByMerchantIDRow, error) {
	var res []*db.GetTransactionsByMerchantIDRow

	err := r.db.SelectContext(ctx, &res, getTransactionsByMerchantID, int32(merchant_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFoundResponse("transaction").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return res, nil
}

func (r *transactionQueryRepository) FindAllTransactionsByMerchantId(ctx context.Context, merchantID int, search string, limit, offset int) ([]*db.GetTransactionsRow, error) {
	var res []*db.GetTransactionsRow

	err := r.db.SelectContext(ctx, &res, getTransactionsByMerchantIDPaginated, int32(merchantID), search, int32(limit), int32(offset))
	if err != nil {
		return nil, sharedErrors.ErrFailed("find all transactions by merchant id").WithInternal(err)
	}

	return res, nil
}
