// Package db holds the database row models for the saldo service.
// Hand-written (migrated from sqlc generation); used with jmoiron/sqlx.
package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type BalanceLedger struct {
	EntryID       int64              `json:"entry_id" db:"entry_id"`
	OperationID   string             `json:"operation_id" db:"operation_id"`
	CardNumber    string             `json:"card_number" db:"card_number"`
	Direction     string             `json:"direction" db:"direction"`
	Amount        int64              `json:"amount" db:"amount"`
	BalanceBefore int64              `json:"balance_before" db:"balance_before"`
	BalanceAfter  int64              `json:"balance_after" db:"balance_after"`
	SourceType    string             `json:"source_type" db:"source_type"`
	SourceID      *string            `json:"source_id" db:"source_id"`
	CreatedAt     pgtype.Timestamptz `json:"created_at" db:"created_at"`
	Delta         int64              `json:"delta" db:"delta"`
	Note          *string            `json:"note" db:"note"`
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

type ReconciliationQueue struct {
	QueueID               int64              `json:"queue_id" db:"queue_id"`
	SaldoID               int32              `json:"saldo_id" db:"saldo_id"`
	CardNumber            string             `json:"card_number" db:"card_number"`
	CurrentBalance        int64              `json:"current_balance" db:"current_balance"`
	LedgerBalance         int64              `json:"ledger_balance" db:"ledger_balance"`
	Difference            int64              `json:"difference" db:"difference"`
	LedgerEntries         int64              `json:"ledger_entries" db:"ledger_entries"`
	Status                string             `json:"status" db:"status"`
	MismatchCount         int64              `json:"mismatch_count" db:"mismatch_count"`
	FirstSeenAt           pgtype.Timestamptz `json:"first_seen_at" db:"first_seen_at"`
	LastSeenAt            pgtype.Timestamptz `json:"last_seen_at" db:"last_seen_at"`
	ResolvedAt            pgtype.Timestamptz `json:"resolved_at" db:"resolved_at"`
	ResolutionOperationID *string            `json:"resolution_operation_id" db:"resolution_operation_id"`
	ResolutionNote        *string            `json:"resolution_note" db:"resolution_note"`
}

type Role struct {
	RoleID    int32            `json:"role_id" db:"role_id"`
	RoleName  string           `json:"role_name" db:"role_name"`
	CreatedAt pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}

type Saldo struct {
	SaldoID        int32            `json:"saldo_id" db:"saldo_id"`
	CardNumber     string           `json:"card_number" db:"card_number"`
	TotalBalance   int64            `json:"total_balance" db:"total_balance"`
	WithdrawAmount *int64           `json:"withdraw_amount" db:"withdraw_amount"`
	WithdrawTime   pgtype.Timestamp `json:"withdraw_time" db:"withdraw_time"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt      pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}

type UserRole struct {
	UserRoleID int32            `json:"user_role_id" db:"user_role_id"`
	UserID     int32            `json:"user_id" db:"user_id"`
	RoleID     int32            `json:"role_id" db:"role_id"`
	CreatedAt  pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt  pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt  pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}
