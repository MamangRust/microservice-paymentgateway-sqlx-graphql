package repository

import (
	"context"
	"math"

	"github.com/jmoiron/sqlx"

	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/google/uuid"
)

const insertCardPayment = `INSERT INTO card_payments (
    payment_uuid, card_number, billing_id, amount, payment_channel, reference_id
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING payment_id, payment_uuid, card_number, billing_id, amount,
          payment_channel, reference_id, status, created_at, updated_at
`

const getCardPaymentsByCardNumber = `SELECT payment_id, payment_uuid, card_number, billing_id, amount,
       payment_channel, reference_id, status, created_at, updated_at,
       COUNT(*) OVER () AS total_count
FROM card_payments
WHERE card_number = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3
`

const countCardPayments = `SELECT COUNT(*)::INT AS total_count
FROM card_payments
WHERE card_number = $1
`

type cardPaymentRepository struct {
	db *sqlx.DB
}

func NewCardPaymentRepository(q *sqlx.DB) CardPaymentRepository {
	return &cardPaymentRepository{db: q}
}

func (r *cardPaymentRepository) PostPayment(ctx context.Context, req *requests.PostPaymentRequest) (*db.CardPayment, error) {
	var billingID *int32
	if req.BillingID != nil {
		billingID = new(int32)
		*billingID = int32(*req.BillingID)
	}

	var res db.CardPayment
	err := r.db.GetContext(ctx, &res, insertCardPayment,
		uuid.New().String(),
		req.CardNumber,
		billingID,
		req.Amount,
		req.PaymentChannel,
		req.ReferenceID,
	)
	if err != nil {
		return nil, sharedErrors.ErrFailed("create card payment").WithInternal(err)
	}
	return &res, nil
}

func (r *cardPaymentRepository) GetPaymentHistory(ctx context.Context, cardNumber string, page, pageSize int) ([]*db.GetCardPaymentsByCardNumberRow, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	var res []*db.GetCardPaymentsByCardNumberRow
	err := r.db.SelectContext(ctx, &res, getCardPaymentsByCardNumber,
		cardNumber,
		int32(pageSize),
		int32(math.Max(0, float64((page-1)*pageSize))),
	)
	if err != nil {
		return nil, sharedErrors.ErrFailed("get card payments").WithInternal(err)
	}
	return res, nil
}

func (r *cardPaymentRepository) CountPayments(ctx context.Context, cardNumber string) (int, error) {
	var count int32
	err := r.db.GetContext(ctx, &count, countCardPayments, cardNumber)
	if err != nil {
		return 0, sharedErrors.ErrFailed("count card payments").WithInternal(err)
	}
	return int(count), nil
}
