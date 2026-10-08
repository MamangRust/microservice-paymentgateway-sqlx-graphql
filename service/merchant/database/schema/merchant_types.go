// Hand-written package db — Row structs migrated from sqlc-generated
// sources (merchant.sql, merchant_document.sql). db tags added for sqlx
// scanning; json tags preserved. Params structs were dropped in favor of
// positional arguments.

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateMerchantDocumentRow struct {
	DocumentID   int32            `db:"document_id" json:"document_id"`
	MerchantID   int32            `db:"merchant_id" json:"merchant_id"`
	DocumentType string           `db:"document_type" json:"document_type"`
	DocumentUrl  string           `db:"document_url" json:"document_url"`
	Status       string           `db:"status" json:"status"`
	Note         *string          `db:"note" json:"note"`
	UploadedAt   pgtype.Timestamp `db:"uploaded_at" json:"uploaded_at"`
	CreatedAt    pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type CreateMerchantRow struct {
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	MerchantNo pgtype.UUID      `db:"merchant_no" json:"merchant_no"`
	Name       string           `db:"name" json:"name"`
	ApiKey     string           `db:"api_key" json:"api_key"`
	UserID     int32            `db:"user_id" json:"user_id"`
	Status     string           `db:"status" json:"status"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetActiveMerchantDocumentsRow struct {
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
	TotalCount   int64            `db:"total_count" json:"total_count"`
}

type GetActiveMerchantsRow struct {
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	MerchantNo pgtype.UUID      `db:"merchant_no" json:"merchant_no"`
	Name       string           `db:"name" json:"name"`
	ApiKey     string           `db:"api_key" json:"api_key"`
	UserID     int32            `db:"user_id" json:"user_id"`
	Status     string           `db:"status" json:"status"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type GetMerchantByApiKeyRow struct {
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	MerchantNo pgtype.UUID      `db:"merchant_no" json:"merchant_no"`
	Name       string           `db:"name" json:"name"`
	ApiKey     string           `db:"api_key" json:"api_key"`
	UserID     int32            `db:"user_id" json:"user_id"`
	Status     string           `db:"status" json:"status"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetMerchantByIDRow struct {
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	MerchantNo pgtype.UUID      `db:"merchant_no" json:"merchant_no"`
	Name       string           `db:"name" json:"name"`
	ApiKey     string           `db:"api_key" json:"api_key"`
	UserID     int32            `db:"user_id" json:"user_id"`
	Status     string           `db:"status" json:"status"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetMerchantByNameRow struct {
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	MerchantNo pgtype.UUID      `db:"merchant_no" json:"merchant_no"`
	Name       string           `db:"name" json:"name"`
	ApiKey     string           `db:"api_key" json:"api_key"`
	UserID     int32            `db:"user_id" json:"user_id"`
	Status     string           `db:"status" json:"status"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetMerchantDocumentRow struct {
	DocumentID   int32            `db:"document_id" json:"document_id"`
	MerchantID   int32            `db:"merchant_id" json:"merchant_id"`
	DocumentType string           `db:"document_type" json:"document_type"`
	DocumentUrl  string           `db:"document_url" json:"document_url"`
	Status       string           `db:"status" json:"status"`
	Note         *string          `db:"note" json:"note"`
	UploadedAt   pgtype.Timestamp `db:"uploaded_at" json:"uploaded_at"`
	CreatedAt    pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetMerchantDocumentsRow struct {
	DocumentID   int32            `db:"document_id" json:"document_id"`
	MerchantID   int32            `db:"merchant_id" json:"merchant_id"`
	DocumentType string           `db:"document_type" json:"document_type"`
	DocumentUrl  string           `db:"document_url" json:"document_url"`
	Status       string           `db:"status" json:"status"`
	Note         *string          `db:"note" json:"note"`
	UploadedAt   pgtype.Timestamp `db:"uploaded_at" json:"uploaded_at"`
	CreatedAt    pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	TotalCount   int64            `db:"total_count" json:"total_count"`
}

type GetMerchantsByUserIDRow struct {
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	MerchantNo pgtype.UUID      `db:"merchant_no" json:"merchant_no"`
	Name       string           `db:"name" json:"name"`
	ApiKey     string           `db:"api_key" json:"api_key"`
	UserID     int32            `db:"user_id" json:"user_id"`
	Status     string           `db:"status" json:"status"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetMerchantsRow struct {
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	MerchantNo pgtype.UUID      `db:"merchant_no" json:"merchant_no"`
	Name       string           `db:"name" json:"name"`
	ApiKey     string           `db:"api_key" json:"api_key"`
	UserID     int32            `db:"user_id" json:"user_id"`
	Status     string           `db:"status" json:"status"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type GetTrashedMerchantDocumentsRow struct {
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
	TotalCount   int64            `db:"total_count" json:"total_count"`
}

type GetTrashedMerchantsRow struct {
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	MerchantNo pgtype.UUID      `db:"merchant_no" json:"merchant_no"`
	Name       string           `db:"name" json:"name"`
	ApiKey     string           `db:"api_key" json:"api_key"`
	UserID     int32            `db:"user_id" json:"user_id"`
	Status     string           `db:"status" json:"status"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type UpdateMerchantDocumentRow struct {
	DocumentID   int32            `db:"document_id" json:"document_id"`
	MerchantID   int32            `db:"merchant_id" json:"merchant_id"`
	DocumentType string           `db:"document_type" json:"document_type"`
	DocumentUrl  string           `db:"document_url" json:"document_url"`
	Status       string           `db:"status" json:"status"`
	Note         *string          `db:"note" json:"note"`
	UploadedAt   pgtype.Timestamp `db:"uploaded_at" json:"uploaded_at"`
	CreatedAt    pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type UpdateMerchantDocumentStatusRow struct {
	DocumentID   int32            `db:"document_id" json:"document_id"`
	MerchantID   int32            `db:"merchant_id" json:"merchant_id"`
	DocumentType string           `db:"document_type" json:"document_type"`
	DocumentUrl  string           `db:"document_url" json:"document_url"`
	Status       string           `db:"status" json:"status"`
	Note         *string          `db:"note" json:"note"`
	UploadedAt   pgtype.Timestamp `db:"uploaded_at" json:"uploaded_at"`
	CreatedAt    pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type UpdateMerchantRow struct {
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	MerchantNo pgtype.UUID      `db:"merchant_no" json:"merchant_no"`
	Name       string           `db:"name" json:"name"`
	ApiKey     string           `db:"api_key" json:"api_key"`
	UserID     int32            `db:"user_id" json:"user_id"`
	Status     string           `db:"status" json:"status"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type UpdateMerchantStatusRow struct {
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	MerchantNo pgtype.UUID      `db:"merchant_no" json:"merchant_no"`
	Name       string           `db:"name" json:"name"`
	ApiKey     string           `db:"api_key" json:"api_key"`
	UserID     int32            `db:"user_id" json:"user_id"`
	Status     string           `db:"status" json:"status"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}
