package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/database/schema"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
)

const getBillingCyclesByCardNumber = `SELECT billing_id, card_number, cycle_start, cycle_end, amount_due, due_date,
       status, created_at, updated_at
FROM billing_cycles
WHERE card_number = $1
ORDER BY cycle_start DESC
`

const insertBillingCycle = `INSERT INTO billing_cycles (
    card_number, cycle_start, cycle_end, amount_due, due_date, status
)
VALUES ($1, $2, $3, $4, $5, 'unpaid')
RETURNING billing_id, card_number, cycle_start, cycle_end, amount_due, due_date,
          status, created_at, updated_at
`

// CreateBillingCycles creates one statement per active card for the period.
// The unique period index makes retries and concurrent schedulers safe.
const createBillingCycles = `INSERT INTO billing_cycles (
    card_number, cycle_start, cycle_end, amount_due, due_date, status
)
SELECT card_number, $1, $2, outstanding_balance, $3, 'unpaid'
FROM cards
WHERE deleted_at IS NULL
ON CONFLICT (card_number, cycle_start) DO NOTHING
RETURNING billing_id, card_number, cycle_start, cycle_end, amount_due, due_date,
          status, created_at, updated_at
`

const updateBillingCycleStatus = `UPDATE billing_cycles
SET status = $2, updated_at = CURRENT_TIMESTAMP
WHERE billing_id = $1
RETURNING billing_id, card_number, cycle_start, cycle_end, amount_due, due_date,
          status, created_at, updated_at
`

type billingCycleRepository struct {
	db *sqlx.DB
}

func NewBillingCycleRepository(q *sqlx.DB) BillingCycleRepository {
	return &billingCycleRepository{db: q}
}

func (r *billingCycleRepository) GetBillingCyclesByCardNumber(ctx context.Context, cardNumber string) ([]*db.BillingCycle, error) {
	var cycles []*db.BillingCycle
	err := r.db.SelectContext(ctx, &cycles, getBillingCyclesByCardNumber, cardNumber)
	if err != nil {
		return nil, sharedErrors.ErrFailed("get billing cycles").WithInternal(err)
	}
	return cycles, nil
}

func (r *billingCycleRepository) CreateBillingCycles(ctx context.Context, cycleStart, cycleEnd, dueDate time.Time) (int, error) {
	var created []*db.BillingCycle
	err := r.db.SelectContext(ctx, &created, createBillingCycles, cycleStart, cycleEnd, dueDate)
	if err != nil {
		return 0, sharedErrors.ErrFailed("create billing cycles").WithInternal(err)
	}
	return len(created), nil
}
