// Hand-written package db — model structs migrated from sqlc-generated
// sources. Struct shapes and json tags preserved; db tags added for sqlx.

package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Merchant struct {
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	MerchantNo pgtype.UUID      `db:"merchant_no" json:"merchant_no"`
	Name       string           `db:"name" json:"name"`
	ApiKey     string           `db:"api_key" json:"api_key"`
	UserID     int32            `db:"user_id" json:"user_id"`
	Status     string           `db:"status" json:"status"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}

type MerchantDocument struct {
	DocumentID   int32            `db:"document_id" json:"document_id"`
	MerchantID   int32            `db:"merchant_id" json:"merchant_id"`
	DocumentType string           `db:"document_type" json:"document_type"`
	DocumentUrl  string           `db:"document_url" json:"document_url"`
	Status       string           `db:"status" json:"status"`
	Note         *string          `db:"note" json:"note"`
	UploadedAt   pgtype.Timestamp `db:"uploaded_at" json:"uploaded_at"`
	CreatedAt    pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt    pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}

type Role struct {
	RoleID    int32            `db:"role_id" json:"role_id"`
	RoleName  string           `db:"role_name" json:"role_name"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}

type Transaction struct {
	TransactionID   int32            `db:"transaction_id" json:"transaction_id"`
	TransactionNo   pgtype.UUID      `db:"transaction_no" json:"transaction_no"`
	CardNumber      string           `db:"card_number" json:"card_number"`
	Amount          int64            `db:"amount" json:"amount"`
	PaymentMethod   string           `db:"payment_method" json:"payment_method"`
	MerchantID      int32            `db:"merchant_id" json:"merchant_id"`
	TransactionTime time.Time        `db:"transaction_time" json:"transaction_time"`
	Status          string           `db:"status" json:"status"`
	CreatedAt       pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt       pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt       pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}

type UserRole struct {
	UserRoleID int32            `db:"user_role_id" json:"user_role_id"`
	UserID     int32            `db:"user_id" json:"user_id"`
	RoleID     int32            `db:"role_id" json:"role_id"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}
