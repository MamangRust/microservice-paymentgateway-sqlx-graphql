// Row/Params types still referenced by repositories and services.
// Hand-written (migrated from sqlc generation); used with jmoiron/sqlx.
package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type CreateWithdrawRow struct {
	WithdrawID     int32            `json:"withdraw_id" db:"withdraw_id"`
	WithdrawNo     pgtype.UUID      `json:"withdraw_no" db:"withdraw_no"`
	CardNumber     string           `json:"card_number" db:"card_number"`
	WithdrawAmount int64            `json:"withdraw_amount" db:"withdraw_amount"`
	WithdrawTime   time.Time        `json:"withdraw_time" db:"withdraw_time"`
	Status         string           `json:"status" db:"status"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type UpdateWithdrawRow struct {
	WithdrawID     int32            `json:"withdraw_id" db:"withdraw_id"`
	WithdrawNo     pgtype.UUID      `json:"withdraw_no" db:"withdraw_no"`
	CardNumber     string           `json:"card_number" db:"card_number"`
	WithdrawAmount int64            `json:"withdraw_amount" db:"withdraw_amount"`
	WithdrawTime   time.Time        `json:"withdraw_time" db:"withdraw_time"`
	Status         string           `json:"status" db:"status"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type UpdateWithdrawStatusRow struct {
	WithdrawID     int32            `json:"withdraw_id" db:"withdraw_id"`
	WithdrawNo     pgtype.UUID      `json:"withdraw_no" db:"withdraw_no"`
	CardNumber     string           `json:"card_number" db:"card_number"`
	WithdrawAmount int64            `json:"withdraw_amount" db:"withdraw_amount"`
	WithdrawTime   time.Time        `json:"withdraw_time" db:"withdraw_time"`
	Status         string           `json:"status" db:"status"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type GetWithdrawsRow struct {
	WithdrawID     int32            `json:"withdraw_id" db:"withdraw_id"`
	WithdrawNo     pgtype.UUID      `json:"withdraw_no" db:"withdraw_no"`
	CardNumber     string           `json:"card_number" db:"card_number"`
	WithdrawAmount int64            `json:"withdraw_amount" db:"withdraw_amount"`
	WithdrawTime   time.Time        `json:"withdraw_time" db:"withdraw_time"`
	Status         string           `json:"status" db:"status"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	TotalCount     int64            `json:"total_count" db:"total_count"`
}

type GetActiveWithdrawsRow struct {
	WithdrawID     int32            `json:"withdraw_id" db:"withdraw_id"`
	WithdrawNo     pgtype.UUID      `json:"withdraw_no" db:"withdraw_no"`
	CardNumber     string           `json:"card_number" db:"card_number"`
	WithdrawAmount int64            `json:"withdraw_amount" db:"withdraw_amount"`
	WithdrawTime   time.Time        `json:"withdraw_time" db:"withdraw_time"`
	Status         string           `json:"status" db:"status"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt      pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
	TotalCount     int64            `json:"total_count" db:"total_count"`
}

type GetTrashedWithdrawsRow struct {
	WithdrawID     int32            `json:"withdraw_id" db:"withdraw_id"`
	WithdrawNo     pgtype.UUID      `json:"withdraw_no" db:"withdraw_no"`
	CardNumber     string           `json:"card_number" db:"card_number"`
	WithdrawAmount int64            `json:"withdraw_amount" db:"withdraw_amount"`
	WithdrawTime   time.Time        `json:"withdraw_time" db:"withdraw_time"`
	Status         string           `json:"status" db:"status"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt      pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
	TotalCount     int64            `json:"total_count" db:"total_count"`
}

type GetWithdrawByIDRow struct {
	WithdrawID     int32            `json:"withdraw_id" db:"withdraw_id"`
	WithdrawNo     pgtype.UUID      `json:"withdraw_no" db:"withdraw_no"`
	CardNumber     string           `json:"card_number" db:"card_number"`
	WithdrawAmount int64            `json:"withdraw_amount" db:"withdraw_amount"`
	WithdrawTime   time.Time        `json:"withdraw_time" db:"withdraw_time"`
	Status         string           `json:"status" db:"status"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type GetWithdrawsByCardNumberRow struct {
	WithdrawID     int32            `json:"withdraw_id" db:"withdraw_id"`
	WithdrawNo     pgtype.UUID      `json:"withdraw_no" db:"withdraw_no"`
	CardNumber     string           `json:"card_number" db:"card_number"`
	WithdrawAmount int64            `json:"withdraw_amount" db:"withdraw_amount"`
	WithdrawTime   time.Time        `json:"withdraw_time" db:"withdraw_time"`
	Status         string           `json:"status" db:"status"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	TotalCount     int64            `json:"total_count" db:"total_count"`
}
