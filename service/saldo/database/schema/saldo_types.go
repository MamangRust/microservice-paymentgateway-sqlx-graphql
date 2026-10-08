// Row/Params types still referenced by repositories, services, and pkg/adapter.
// Hand-written (migrated from sqlc generation); used with jmoiron/sqlx.
package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateSaldoRow struct {
	SaldoID        int32            `json:"saldo_id" db:"saldo_id"`
	CardNumber     string           `json:"card_number" db:"card_number"`
	TotalBalance   int64            `json:"total_balance" db:"total_balance"`
	WithdrawAmount *int64           `json:"withdraw_amount" db:"withdraw_amount"`
	WithdrawTime   pgtype.Timestamp `json:"withdraw_time" db:"withdraw_time"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type CreditSaldoRow struct {
	SaldoID        int32            `json:"saldo_id" db:"saldo_id"`
	CardNumber     string           `json:"card_number" db:"card_number"`
	TotalBalance   int64            `json:"total_balance" db:"total_balance"`
	WithdrawAmount *int64           `json:"withdraw_amount" db:"withdraw_amount"`
	WithdrawTime   pgtype.Timestamp `json:"withdraw_time" db:"withdraw_time"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type DebitSaldoRow struct {
	SaldoID        int32            `json:"saldo_id" db:"saldo_id"`
	CardNumber     string           `json:"card_number" db:"card_number"`
	TotalBalance   int64            `json:"total_balance" db:"total_balance"`
	WithdrawAmount *int64           `json:"withdraw_amount" db:"withdraw_amount"`
	WithdrawTime   pgtype.Timestamp `json:"withdraw_time" db:"withdraw_time"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type GetActiveSaldosRow struct {
	SaldoID        int32            `json:"saldo_id" db:"saldo_id"`
	CardNumber     string           `json:"card_number" db:"card_number"`
	TotalBalance   int64            `json:"total_balance" db:"total_balance"`
	WithdrawAmount *int64           `json:"withdraw_amount" db:"withdraw_amount"`
	WithdrawTime   pgtype.Timestamp `json:"withdraw_time" db:"withdraw_time"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt      pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
	TotalCount     int64            `json:"total_count" db:"total_count"`
}

type GetSaldoByIDRow struct {
	SaldoID        int32            `json:"saldo_id" db:"saldo_id"`
	CardNumber     string           `json:"card_number" db:"card_number"`
	TotalBalance   int64            `json:"total_balance" db:"total_balance"`
	WithdrawAmount *int64           `json:"withdraw_amount" db:"withdraw_amount"`
	WithdrawTime   pgtype.Timestamp `json:"withdraw_time" db:"withdraw_time"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type GetSaldosRow struct {
	SaldoID        int32            `json:"saldo_id" db:"saldo_id"`
	CardNumber     string           `json:"card_number" db:"card_number"`
	TotalBalance   int64            `json:"total_balance" db:"total_balance"`
	WithdrawAmount *int64           `json:"withdraw_amount" db:"withdraw_amount"`
	WithdrawTime   pgtype.Timestamp `json:"withdraw_time" db:"withdraw_time"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	TotalCount     int64            `json:"total_count" db:"total_count"`
}

type GetTrashedSaldosRow struct {
	SaldoID        int32            `json:"saldo_id" db:"saldo_id"`
	CardNumber     string           `json:"card_number" db:"card_number"`
	TotalBalance   int64            `json:"total_balance" db:"total_balance"`
	WithdrawAmount *int64           `json:"withdraw_amount" db:"withdraw_amount"`
	WithdrawTime   pgtype.Timestamp `json:"withdraw_time" db:"withdraw_time"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt      pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
	TotalCount     int64            `json:"total_count" db:"total_count"`
}

type UpdateSaldoRow struct {
	SaldoID        int32            `json:"saldo_id" db:"saldo_id"`
	CardNumber     string           `json:"card_number" db:"card_number"`
	TotalBalance   int64            `json:"total_balance" db:"total_balance"`
	WithdrawAmount *int64           `json:"withdraw_amount" db:"withdraw_amount"`
	WithdrawTime   pgtype.Timestamp `json:"withdraw_time" db:"withdraw_time"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type UpdateSaldoBalanceRow struct {
	SaldoID        int32            `json:"saldo_id" db:"saldo_id"`
	CardNumber     string           `json:"card_number" db:"card_number"`
	TotalBalance   int64            `json:"total_balance" db:"total_balance"`
	WithdrawAmount *int64           `json:"withdraw_amount" db:"withdraw_amount"`
	WithdrawTime   pgtype.Timestamp `json:"withdraw_time" db:"withdraw_time"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type UpdateSaldoWithdrawRow struct {
	SaldoID        int32            `json:"saldo_id" db:"saldo_id"`
	CardNumber     string           `json:"card_number" db:"card_number"`
	TotalBalance   int64            `json:"total_balance" db:"total_balance"`
	WithdrawAmount *int64           `json:"withdraw_amount" db:"withdraw_amount"`
	WithdrawTime   pgtype.Timestamp `json:"withdraw_time" db:"withdraw_time"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type UpdateSaldoWithdrawMetadataRow struct {
	SaldoID        int32            `json:"saldo_id" db:"saldo_id"`
	CardNumber     string           `json:"card_number" db:"card_number"`
	TotalBalance   int64            `json:"total_balance" db:"total_balance"`
	WithdrawAmount *int64           `json:"withdraw_amount" db:"withdraw_amount"`
	WithdrawTime   pgtype.Timestamp `json:"withdraw_time" db:"withdraw_time"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}
