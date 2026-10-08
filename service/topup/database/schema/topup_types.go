package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type CreateTopupRow struct {
	TopupID     int32            `json:"topup_id" db:"topup_id"`
	TopupNo     pgtype.UUID      `json:"topup_no" db:"topup_no"`
	CardNumber  string           `json:"card_number" db:"card_number"`
	TopupAmount int64            `json:"topup_amount" db:"topup_amount"`
	TopupMethod string           `json:"topup_method" db:"topup_method"`
	TopupTime   time.Time        `json:"topup_time" db:"topup_time"`
	Status      string           `json:"status" db:"status"`
	CreatedAt   pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt   pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type GetActiveTopupsRow struct {
	TopupID     int32            `json:"topup_id" db:"topup_id"`
	TopupNo     pgtype.UUID      `json:"topup_no" db:"topup_no"`
	CardNumber  string           `json:"card_number" db:"card_number"`
	TopupAmount int64            `json:"topup_amount" db:"topup_amount"`
	TopupMethod string           `json:"topup_method" db:"topup_method"`
	TopupTime   time.Time        `json:"topup_time" db:"topup_time"`
	Status      string           `json:"status" db:"status"`
	CreatedAt   pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt   pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt   pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
	TotalCount  int64            `json:"total_count" db:"total_count"`
}

type GetTopupByIDRow struct {
	TopupID     int32            `json:"topup_id" db:"topup_id"`
	TopupNo     pgtype.UUID      `json:"topup_no" db:"topup_no"`
	CardNumber  string           `json:"card_number" db:"card_number"`
	TopupAmount int64            `json:"topup_amount" db:"topup_amount"`
	TopupMethod string           `json:"topup_method" db:"topup_method"`
	TopupTime   time.Time        `json:"topup_time" db:"topup_time"`
	Status      string           `json:"status" db:"status"`
	CreatedAt   pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt   pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type GetTopupsByCardNumberRow struct {
	TopupID     int32            `json:"topup_id" db:"topup_id"`
	TopupNo     pgtype.UUID      `json:"topup_no" db:"topup_no"`
	CardNumber  string           `json:"card_number" db:"card_number"`
	TopupAmount int64            `json:"topup_amount" db:"topup_amount"`
	TopupMethod string           `json:"topup_method" db:"topup_method"`
	TopupTime   time.Time        `json:"topup_time" db:"topup_time"`
	Status      string           `json:"status" db:"status"`
	CreatedAt   pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt   pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	TotalCount  int64            `json:"total_count" db:"total_count"`
}

type GetTopupsRow struct {
	TopupID     int32            `json:"topup_id" db:"topup_id"`
	TopupNo     pgtype.UUID      `json:"topup_no" db:"topup_no"`
	CardNumber  string           `json:"card_number" db:"card_number"`
	TopupAmount int64            `json:"topup_amount" db:"topup_amount"`
	TopupMethod string           `json:"topup_method" db:"topup_method"`
	TopupTime   time.Time        `json:"topup_time" db:"topup_time"`
	Status      string           `json:"status" db:"status"`
	CreatedAt   pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt   pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	TotalCount  int64            `json:"total_count" db:"total_count"`
}

type GetTrashedTopupsRow struct {
	TopupID     int32            `json:"topup_id" db:"topup_id"`
	TopupNo     pgtype.UUID      `json:"topup_no" db:"topup_no"`
	CardNumber  string           `json:"card_number" db:"card_number"`
	TopupAmount int64            `json:"topup_amount" db:"topup_amount"`
	TopupMethod string           `json:"topup_method" db:"topup_method"`
	TopupTime   time.Time        `json:"topup_time" db:"topup_time"`
	Status      string           `json:"status" db:"status"`
	CreatedAt   pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt   pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt   pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
	TotalCount  int64            `json:"total_count" db:"total_count"`
}

type UpdateTopupAmountRow struct {
	TopupID     int32            `json:"topup_id" db:"topup_id"`
	TopupNo     pgtype.UUID      `json:"topup_no" db:"topup_no"`
	CardNumber  string           `json:"card_number" db:"card_number"`
	TopupAmount int64            `json:"topup_amount" db:"topup_amount"`
	TopupMethod string           `json:"topup_method" db:"topup_method"`
	TopupTime   time.Time        `json:"topup_time" db:"topup_time"`
	Status      string           `json:"status" db:"status"`
	CreatedAt   pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt   pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type UpdateTopupRow struct {
	TopupID     int32            `json:"topup_id" db:"topup_id"`
	TopupNo     pgtype.UUID      `json:"topup_no" db:"topup_no"`
	CardNumber  string           `json:"card_number" db:"card_number"`
	TopupAmount int64            `json:"topup_amount" db:"topup_amount"`
	TopupMethod string           `json:"topup_method" db:"topup_method"`
	TopupTime   time.Time        `json:"topup_time" db:"topup_time"`
	Status      string           `json:"status" db:"status"`
	CreatedAt   pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt   pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type UpdateTopupStatusRow struct {
	TopupID     int32            `json:"topup_id" db:"topup_id"`
	TopupNo     pgtype.UUID      `json:"topup_no" db:"topup_no"`
	CardNumber  string           `json:"card_number" db:"card_number"`
	TopupAmount int64            `json:"topup_amount" db:"topup_amount"`
	TopupMethod string           `json:"topup_method" db:"topup_method"`
	TopupTime   time.Time        `json:"topup_time" db:"topup_time"`
	Status      string           `json:"status" db:"status"`
	CreatedAt   pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt   pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}
