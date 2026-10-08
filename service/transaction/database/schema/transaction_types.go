package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type GetTransactionsRow struct {
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
	TotalCount      int64            `json:"total_count" db:"total_count"`
}

type GetActiveTransactionsRow struct {
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
	TotalCount      int64            `json:"total_count" db:"total_count"`
}

type GetTrashedTransactionsRow struct {
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
	TotalCount      int64            `json:"total_count" db:"total_count"`
}

type GetTransactionByIDRow struct {
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
}

type GetTransactionsByCardNumberRow struct {
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
	TotalCount      int64            `json:"total_count" db:"total_count"`
}

type GetTransactionsByMerchantIDRow struct {
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
}

type CreateTransactionRow struct {
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
}

type UpdateTransactionRow struct {
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
}

type UpdateTransactionStatusRow struct {
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
}
