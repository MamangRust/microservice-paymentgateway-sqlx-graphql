// Package db holds the database row models for the withdraw service.
// Hand-written (migrated from sqlc generation); used with jmoiron/sqlx.
package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type IdempotencyRecord struct {
	ID              int64       `json:"id" db:"id"`
	Scope           string      `json:"scope" db:"scope"`
	IdempotencyKey  string      `json:"idempotency_key" db:"idempotency_key"`
	RequestHash     string      `json:"request_hash" db:"request_hash"`
	OperationID     pgtype.UUID `json:"operation_id" db:"operation_id"`
	Status          string      `json:"status" db:"status"`
	ResponsePayload []byte      `json:"response_payload" db:"response_payload"`
	ResourceID      *int32      `json:"resource_id" db:"resource_id"`
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at" db:"updated_at"`
	ExpiresAt       time.Time   `json:"expires_at" db:"expires_at"`
}

type Withdraw struct {
	WithdrawID     int32            `json:"withdraw_id" db:"withdraw_id"`
	WithdrawNo     pgtype.UUID      `json:"withdraw_no" db:"withdraw_no"`
	CardNumber     string           `json:"card_number" db:"card_number"`
	WithdrawAmount int64            `json:"withdraw_amount" db:"withdraw_amount"`
	WithdrawTime   time.Time        `json:"withdraw_time" db:"withdraw_time"`
	Status         string           `json:"status" db:"status"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt      pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}
