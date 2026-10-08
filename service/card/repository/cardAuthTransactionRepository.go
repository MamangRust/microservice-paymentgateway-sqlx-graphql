package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jmoiron/sqlx"

	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
)

const insertAuthTransaction = `INSERT INTO card_auth_transactions (
    txn_id, card_number, merchant_id, amount, currency, mcc,
    pos_entry_mode, idempotency_key
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING auth_id, txn_id, card_number, merchant_id, amount, currency, mcc,
          pos_entry_mode, status, idempotency_key, risk_score, created_at, updated_at
`

const approveAuthTransaction = `UPDATE card_auth_transactions
SET status = 'approved', updated_at = CURRENT_TIMESTAMP
WHERE txn_id = $1 AND status = 'pending'
RETURNING auth_id, txn_id, card_number, merchant_id, amount, currency, mcc,
          pos_entry_mode, status, idempotency_key, risk_score, created_at, updated_at
`

const declineAuthTransaction = `UPDATE card_auth_transactions
SET status = 'declined', updated_at = CURRENT_TIMESTAMP
WHERE txn_id = $1 AND status = 'pending'
RETURNING auth_id, txn_id, card_number, merchant_id, amount, currency, mcc,
          pos_entry_mode, status, idempotency_key, risk_score, created_at, updated_at
`

const reverseAuthTransaction = `UPDATE card_auth_transactions
SET status = 'reversed', updated_at = CURRENT_TIMESTAMP
WHERE txn_id = $1 AND status = 'approved'
RETURNING auth_id, txn_id, card_number, merchant_id, amount, currency, mcc,
          pos_entry_mode, status, idempotency_key, risk_score, created_at, updated_at
`

const getAuthTransactionByIdempotencyKey = `SELECT auth_id, txn_id, card_number, merchant_id, amount, currency, mcc,
       pos_entry_mode, status, idempotency_key, risk_score, created_at, updated_at
FROM card_auth_transactions
WHERE idempotency_key = $1
`

const getAuthTransactionByTxnId = `SELECT auth_id, txn_id, card_number, merchant_id, amount, currency, mcc,
       pos_entry_mode, status, idempotency_key, risk_score, created_at, updated_at
FROM card_auth_transactions
WHERE txn_id = $1
`

const getAuthTransactionsByCardNumber = `SELECT auth_id, txn_id, card_number, merchant_id, amount, currency, mcc,
       pos_entry_mode, status, idempotency_key, risk_score, created_at, updated_at,
       COUNT(*) OVER () AS total_count
FROM card_auth_transactions
WHERE card_number = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3
`

const countRecentAuthByCardNumber = `SELECT COUNT(*)::INT AS total_count
FROM card_auth_transactions
WHERE card_number = $1 AND created_at >= $2
`

const updateAuthTransactionRiskScore = `UPDATE card_auth_transactions
SET risk_score = $2, updated_at = CURRENT_TIMESTAMP
WHERE txn_id = $1
`

type cardAuthTransactionRepository struct {
	db *sqlx.DB
}

func NewCardAuthTransactionRepository(q *sqlx.DB) CardAuthTransactionRepository {
	return &cardAuthTransactionRepository{db: q}
}

func (r *cardAuthTransactionRepository) InsertPending(ctx context.Context, req *requests.AuthorizeCardRequest) (*db.CardAuthTransaction, error) {
	var res db.CardAuthTransaction
	err := r.db.GetContext(ctx, &res, insertAuthTransaction,
		req.IdempotencyKey+"-txn",
		req.CardNumber,
		int32(req.MerchantID),
		req.Amount,
		req.Currency,
		req.Mcc,
		req.PosEntryMode,
		req.IdempotencyKey,
	)
	if err != nil {
		return nil, sharedErrors.ErrFailed("insert authorization transaction").WithInternal(err)
	}
	return &res, nil
}

func (r *cardAuthTransactionRepository) Approve(ctx context.Context, txnID string) (*db.CardAuthTransaction, error) {
	var res db.CardAuthTransaction
	err := r.db.GetContext(ctx, &res, approveAuthTransaction, txnID)
	if err != nil {
		return nil, sharedErrors.ErrFailed("approve authorization transaction").WithInternal(err)
	}
	return &res, nil
}

func (r *cardAuthTransactionRepository) Decline(ctx context.Context, txnID string) (*db.CardAuthTransaction, error) {
	var res db.CardAuthTransaction
	err := r.db.GetContext(ctx, &res, declineAuthTransaction, txnID)
	if err != nil {
		return nil, sharedErrors.ErrFailed("decline authorization transaction").WithInternal(err)
	}
	return &res, nil
}

func (r *cardAuthTransactionRepository) Reverse(ctx context.Context, txnID string) (*db.CardAuthTransaction, error) {
	var res db.CardAuthTransaction
	err := r.db.GetContext(ctx, &res, reverseAuthTransaction, txnID)
	if err != nil {
		return nil, sharedErrors.ErrFailed("reverse authorization transaction").WithInternal(err)
	}
	return &res, nil
}

func (r *cardAuthTransactionRepository) FindByIdempotencyKey(ctx context.Context, key string) (*db.CardAuthTransaction, error) {
	var res db.CardAuthTransaction
	err := r.db.GetContext(ctx, &res, getAuthTransactionByIdempotencyKey, key)
	if err != nil {
		return nil, sharedErrors.ErrFailed("get authorization transaction").WithInternal(err)
	}
	return &res, nil
}

func (r *cardAuthTransactionRepository) FindByTxnID(ctx context.Context, txnID string) (*db.CardAuthTransaction, error) {
	var res db.CardAuthTransaction
	err := r.db.GetContext(ctx, &res, getAuthTransactionByTxnId, txnID)
	if err != nil {
		return nil, sharedErrors.ErrFailed("get authorization transaction").WithInternal(err)
	}
	return &res, nil
}

func (r *cardAuthTransactionRepository) FindByCardNumber(ctx context.Context, cardNumber string, page, pageSize int) ([]*db.GetAuthTransactionsByCardNumberRow, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	var res []*db.GetAuthTransactionsByCardNumberRow
	err := r.db.SelectContext(ctx, &res, getAuthTransactionsByCardNumber,
		cardNumber,
		int32(pageSize),
		int32((page-1)*pageSize),
	)
	if err != nil {
		return nil, sharedErrors.ErrFailed("get authorization transactions").WithInternal(err)
	}
	return res, nil
}

func (r *cardAuthTransactionRepository) CountRecentByCardNumber(ctx context.Context, cardNumber string, since time.Time) (int, error) {
	var count int32
	err := r.db.GetContext(ctx, &count, countRecentAuthByCardNumber, cardNumber, pgtype.Timestamp{Time: since, Valid: true})
	if err != nil {
		return 0, sharedErrors.ErrFailed("count recent authorization transactions").WithInternal(err)
	}
	return int(count), nil
}

func (r *cardAuthTransactionRepository) UpdateRiskScore(ctx context.Context, txnID string, score int) error {
	if _, err := r.db.ExecContext(ctx, updateAuthTransactionRiskScore, txnID, int32(score)); err != nil {
		return sharedErrors.ErrFailed("update authorization risk score").WithInternal(err)
	}
	return nil
}
