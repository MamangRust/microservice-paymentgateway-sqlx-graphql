package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/topup/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/idempotency"
	"github.com/jmoiron/sqlx"
)

// TopupIdempotencyRepository adapts the shared idempotency.Store contract to the
// topup service DB's idempotency_records table.
type TopupIdempotencyRepository struct {
	db *sqlx.DB
}

func NewTopupIdempotencyRepository(db *sqlx.DB) *TopupIdempotencyRepository {
	return &TopupIdempotencyRepository{db: db}
}

func (r *TopupIdempotencyRepository) Claim(ctx context.Context, scope, key, requestHash string) (*idempotency.Record, error) {
	rec, err := db.ClaimIdempotency(ctx, r.db, scope, key, requestHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, idempotency.ErrKeyInUse
		}
		return nil, err
	}
	return mapIdempotencyRecord(rec), nil
}

func (r *TopupIdempotencyRepository) Get(ctx context.Context, scope, key string) (*idempotency.Record, error) {
	rec, err := db.GetIdempotency(ctx, r.db, scope, key)
	if err != nil {
		return nil, err
	}
	return mapIdempotencyRecord(rec), nil
}

func (r *TopupIdempotencyRepository) Complete(ctx context.Context, scope, key, requestHash string, outcome idempotency.Outcome) error {
	return db.CompleteIdempotency(ctx, r.db, db.CompleteIdempotencyParams{
		Scope:           scope,
		IdempotencyKey:  key,
		RequestHash:     requestHash,
		Status:          outcome.Status,
		ResponsePayload: outcome.ResponseJSON,
		ResourceID:      outcome.ResourceID,
	})
}

func (r *TopupIdempotencyRepository) Release(ctx context.Context, scope, key, requestHash string) error {
	return db.ReleaseIdempotency(ctx, r.db, scope, key, requestHash)
}

func mapIdempotencyRecord(rec *db.IdempotencyRecord) *idempotency.Record {
	if rec == nil {
		return nil
	}
	return &idempotency.Record{
		ID:           rec.ID,
		Key:          rec.IdempotencyKey,
		RequestHash:  rec.RequestHash,
		OperationID:  rec.OperationID.String(),
		Status:       rec.Status,
		ResponseJSON: rec.ResponsePayload,
		ResourceID:   rec.ResourceID,
		CreatedAt:    rec.CreatedAt,
		UpdatedAt:    rec.UpdatedAt,
		ExpiresAt:    rec.ExpiresAt,
	}
}
