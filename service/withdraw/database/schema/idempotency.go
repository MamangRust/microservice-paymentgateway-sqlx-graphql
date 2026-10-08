package db

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
)

// IdempotencyRecord comes from models.go (migrated sqlc model). This file only
// adds the handwritten query helpers (ClaimIdempotency, GetIdempotency, ...);
// the operation_id column is typed pgtype.UUID to match the DB; the mapper in
// the command repository converts it to a string for the shared
// idempotency.Record.

// CompleteIdempotencyParams carries the write payload for CompleteIdempotency.
type CompleteIdempotencyParams struct {
	Scope           string `json:"scope"`
	IdempotencyKey  string `json:"idempotency_key"`
	RequestHash     string `json:"request_hash"`
	Status          string `json:"status"`
	ResponsePayload []byte `json:"response_payload"`
	ResourceID      *int32 `json:"resource_id"`
}

const idempotencyColumns = `id, scope, idempotency_key, request_hash, operation_id, status, response_payload, resource_id, created_at, updated_at, expires_at`

// ClaimIdempotency atomically claims a key. sql.ErrNoRows means another request
// already owns the key (and it has not expired) and callers should load it with
// GetIdempotency. Expired records are reclaimed by the same request, so a client
// retry after the 24h window never double-executes against a stale claim.
func ClaimIdempotency(ctx context.Context, q sqlx.QueryerContext, scope, key, requestHash string) (*IdempotencyRecord, error) {
	var record IdempotencyRecord
	err := sqlx.GetContext(ctx, q, &record, `
		INSERT INTO idempotency_records (scope, idempotency_key, request_hash, status)
		VALUES ($1, $2, $3, 'processing')
		ON CONFLICT (scope, idempotency_key) DO UPDATE
			SET request_hash = EXCLUDED.request_hash,
			    status = 'processing',
			    response_payload = NULL,
			    resource_id = NULL,
			    updated_at = current_timestamp,
			    expires_at = current_timestamp + INTERVAL '24 hours'
			WHERE idempotency_records.expires_at <= current_timestamp
		RETURNING `+idempotencyColumns, scope, key, requestHash)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func GetIdempotency(ctx context.Context, q sqlx.QueryerContext, scope, key string) (*IdempotencyRecord, error) {
	var record IdempotencyRecord
	err := sqlx.GetContext(ctx, q, &record, `SELECT `+idempotencyColumns+` FROM idempotency_records WHERE scope=$1 AND idempotency_key=$2 AND expires_at > current_timestamp`, scope, key)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func CompleteIdempotency(ctx context.Context, db sqlx.ExtContext, arg CompleteIdempotencyParams) error {
	_, err := db.ExecContext(ctx, `
		UPDATE idempotency_records
		SET status=$4, response_payload=$5, resource_id=$6, updated_at=current_timestamp
		WHERE scope=$1 AND idempotency_key=$2 AND request_hash=$3`,
		arg.Scope, arg.IdempotencyKey, arg.RequestHash, arg.Status, arg.ResponsePayload, arg.ResourceID)
	return err
}

func ReleaseIdempotency(ctx context.Context, db sqlx.ExtContext, scope, key, requestHash string) error {
	_, err := db.ExecContext(ctx, `DELETE FROM idempotency_records WHERE scope=$1 AND idempotency_key=$2 AND request_hash=$3 AND status='processing'`, scope, key, requestHash)
	return err
}

var errIdempotencyNoRows = errors.New("idempotency record not found")

func IsIdempotencyMiss(err error) bool { return errors.Is(err, errIdempotencyNoRows) }
