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

type OutboxEvent struct {
	ID            int64              `json:"id" db:"id"`
	AggregateType string             `json:"aggregate_type" db:"aggregate_type"`
	AggregateID   string             `json:"aggregate_id" db:"aggregate_id"`
	EventID       string             `json:"event_id" db:"event_id"`
	EventType     string             `json:"event_type" db:"event_type"`
	EventVersion  int32              `json:"event_version" db:"event_version"`
	Payload       []byte             `json:"payload" db:"payload"`
	Status        string             `json:"status" db:"status"`
	Attempts      int32              `json:"attempts" db:"attempts"`
	LastError     *string            `json:"last_error" db:"last_error"`
	AvailableAt   pgtype.Timestamptz `json:"available_at" db:"available_at"`
	CreatedAt     pgtype.Timestamptz `json:"created_at" db:"created_at"`
	PublishedAt   pgtype.Timestamptz `json:"published_at" db:"published_at"`
}

type Role struct {
	RoleID    int32            `json:"role_id" db:"role_id"`
	RoleName  string           `json:"role_name" db:"role_name"`
	CreatedAt pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}

type Transaction struct {
	TransactionID   int32            `json:"transaction_id" db:"transaction_id"`
	TransactionNo   pgtype.UUID      `json:"transaction_no" db:"transaction_no"`
	CardNumber      string           `json:"card_number" db:"card_number"`
	Amount          int64            `json:"amount" db:"amount"`
	PaymentMethod   string           `json:"payment_method" db:"payment_method"`
	MerchantID      int32            `json:"merchant_id" db:"merchant_id"`
	TransactionTime time.Time        `json:"transaction_time" db:"transaction_time"`
	Status          string           `json:"status" db:"status"`
	CreatedAt       pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt       pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt       pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
	OperationID     pgtype.UUID      `json:"operation_id" db:"operation_id"`
	FailureReason   *string          `json:"failure_reason" db:"failure_reason"`
}

type UserRole struct {
	UserRoleID int32            `json:"user_role_id" db:"user_role_id"`
	UserID     int32            `json:"user_id" db:"user_id"`
	RoleID     int32            `json:"role_id" db:"role_id"`
	CreatedAt  pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt  pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt  pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}
