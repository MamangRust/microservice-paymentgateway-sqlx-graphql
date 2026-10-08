// Row types for the card service queries. This file replaces the
// sqlc-generated Row/Params structs: field types and json tags are unchanged
// from the generated version, with sqlx `db` tags added for column mapping.
// Params structs were dropped — queries take positional arguments in the
// exact $n order.
package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateCardRow struct {
	CardID       int32            `json:"card_id" db:"card_id"`
	UserID       int32            `json:"user_id" db:"user_id"`
	CardNumber   string           `json:"card_number" db:"card_number"`
	CardType     string           `json:"card_type" db:"card_type"`
	ExpireDate   pgtype.Date      `json:"expire_date" db:"expire_date"`
	Cvv          string           `json:"cvv" db:"cvv"`
	CardProvider string           `json:"card_provider" db:"card_provider"`
	CreatedAt    pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt    pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type GetCardsRow struct {
	CardID       int32            `json:"card_id" db:"card_id"`
	UserID       int32            `json:"user_id" db:"user_id"`
	CardNumber   string           `json:"card_number" db:"card_number"`
	CardType     string           `json:"card_type" db:"card_type"`
	ExpireDate   pgtype.Date      `json:"expire_date" db:"expire_date"`
	Cvv          string           `json:"cvv" db:"cvv"`
	CardProvider string           `json:"card_provider" db:"card_provider"`
	CreatedAt    pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt    pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	TotalCount   int64            `json:"total_count" db:"total_count"`
}

type GetActiveCardsWithCountRow struct {
	CardID       int32            `json:"card_id" db:"card_id"`
	UserID       int32            `json:"user_id" db:"user_id"`
	CardNumber   string           `json:"card_number" db:"card_number"`
	CardType     string           `json:"card_type" db:"card_type"`
	ExpireDate   pgtype.Date      `json:"expire_date" db:"expire_date"`
	Cvv          string           `json:"cvv" db:"cvv"`
	CardProvider string           `json:"card_provider" db:"card_provider"`
	CreatedAt    pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt    pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt    pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
	TotalCount   int64            `json:"total_count" db:"total_count"`
}

type GetTrashedCardsWithCountRow struct {
	CardID       int32            `json:"card_id" db:"card_id"`
	UserID       int32            `json:"user_id" db:"user_id"`
	CardNumber   string           `json:"card_number" db:"card_number"`
	CardType     string           `json:"card_type" db:"card_type"`
	ExpireDate   pgtype.Date      `json:"expire_date" db:"expire_date"`
	Cvv          string           `json:"cvv" db:"cvv"`
	CardProvider string           `json:"card_provider" db:"card_provider"`
	CreatedAt    pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt    pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt    pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
	TotalCount   int64            `json:"total_count" db:"total_count"`
}

type GetCardByIDRow struct {
	CardID       int32            `json:"card_id" db:"card_id"`
	UserID       int32            `json:"user_id" db:"user_id"`
	CardNumber   string           `json:"card_number" db:"card_number"`
	CardType     string           `json:"card_type" db:"card_type"`
	ExpireDate   pgtype.Date      `json:"expire_date" db:"expire_date"`
	Cvv          string           `json:"cvv" db:"cvv"`
	CardProvider string           `json:"card_provider" db:"card_provider"`
	CreatedAt    pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt    pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type GetCardByUserIDRow struct {
	CardID       int32            `json:"card_id" db:"card_id"`
	UserID       int32            `json:"user_id" db:"user_id"`
	CardNumber   string           `json:"card_number" db:"card_number"`
	CardType     string           `json:"card_type" db:"card_type"`
	ExpireDate   pgtype.Date      `json:"expire_date" db:"expire_date"`
	Cvv          string           `json:"cvv" db:"cvv"`
	CardProvider string           `json:"card_provider" db:"card_provider"`
	CreatedAt    pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt    pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type GetCardByCardNumberRow struct {
	CardID       int32            `json:"card_id" db:"card_id"`
	UserID       int32            `json:"user_id" db:"user_id"`
	CardNumber   string           `json:"card_number" db:"card_number"`
	CardType     string           `json:"card_type" db:"card_type"`
	ExpireDate   pgtype.Date      `json:"expire_date" db:"expire_date"`
	Cvv          string           `json:"cvv" db:"cvv"`
	CardProvider string           `json:"card_provider" db:"card_provider"`
	CreatedAt    pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt    pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type GetUserEmailByCardNumberRow struct {
	CardID       int32            `json:"card_id" db:"card_id"`
	Email        string           `json:"email" db:"email"`
	UserID       int32            `json:"user_id" db:"user_id"`
	CardNumber   string           `json:"card_number" db:"card_number"`
	CardType     string           `json:"card_type" db:"card_type"`
	ExpireDate   pgtype.Date      `json:"expire_date" db:"expire_date"`
	Cvv          string           `json:"cvv" db:"cvv"`
	CardProvider string           `json:"card_provider" db:"card_provider"`
	CreatedAt    pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt    pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type TrashCardRow struct {
	CardID       int32            `json:"card_id" db:"card_id"`
	UserID       int32            `json:"user_id" db:"user_id"`
	CardNumber   string           `json:"card_number" db:"card_number"`
	CardType     string           `json:"card_type" db:"card_type"`
	ExpireDate   pgtype.Date      `json:"expire_date" db:"expire_date"`
	Cvv          string           `json:"cvv" db:"cvv"`
	CardProvider string           `json:"card_provider" db:"card_provider"`
	CreatedAt    pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt    pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt    pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}

type RestoreCardRow struct {
	CardID       int32            `json:"card_id" db:"card_id"`
	UserID       int32            `json:"user_id" db:"user_id"`
	CardNumber   string           `json:"card_number" db:"card_number"`
	CardType     string           `json:"card_type" db:"card_type"`
	ExpireDate   pgtype.Date      `json:"expire_date" db:"expire_date"`
	Cvv          string           `json:"cvv" db:"cvv"`
	CardProvider string           `json:"card_provider" db:"card_provider"`
	CreatedAt    pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt    pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt    pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}

type UpdateCardRow struct {
	CardID       int32            `json:"card_id" db:"card_id"`
	UserID       int32            `json:"user_id" db:"user_id"`
	CardNumber   string           `json:"card_number" db:"card_number"`
	CardType     string           `json:"card_type" db:"card_type"`
	ExpireDate   pgtype.Date      `json:"expire_date" db:"expire_date"`
	Cvv          string           `json:"cvv" db:"cvv"`
	CardProvider string           `json:"card_provider" db:"card_provider"`
	CreatedAt    pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt    pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type ToggleCardStatusRow struct {
	CardID             int32            `json:"card_id" db:"card_id"`
	UserID             int32            `json:"user_id" db:"user_id"`
	CardNumber         string           `json:"card_number" db:"card_number"`
	CardType           string           `json:"card_type" db:"card_type"`
	ExpireDate         pgtype.Date      `json:"expire_date" db:"expire_date"`
	Cvv                string           `json:"cvv" db:"cvv"`
	CardProvider       string           `json:"card_provider" db:"card_provider"`
	Status             string           `json:"status" db:"status"`
	CreditLimit        int32            `json:"credit_limit" db:"credit_limit"`
	OutstandingBalance int32            `json:"outstanding_balance" db:"outstanding_balance"`
	RewardPoints       int32            `json:"reward_points" db:"reward_points"`
	CreatedAt          pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt          pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type UpdateCardStatusRow struct {
	CardID             int32            `json:"card_id" db:"card_id"`
	UserID             int32            `json:"user_id" db:"user_id"`
	CardNumber         string           `json:"card_number" db:"card_number"`
	CardType           string           `json:"card_type" db:"card_type"`
	ExpireDate         pgtype.Date      `json:"expire_date" db:"expire_date"`
	Cvv                string           `json:"cvv" db:"cvv"`
	CardProvider       string           `json:"card_provider" db:"card_provider"`
	Status             string           `json:"status" db:"status"`
	CreditLimit        int32            `json:"credit_limit" db:"credit_limit"`
	OutstandingBalance int32            `json:"outstanding_balance" db:"outstanding_balance"`
	RewardPoints       int32            `json:"reward_points" db:"reward_points"`
	CreatedAt          pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt          pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type UpdateCreditLimitRow struct {
	CardID             int32            `json:"card_id" db:"card_id"`
	UserID             int32            `json:"user_id" db:"user_id"`
	CardNumber         string           `json:"card_number" db:"card_number"`
	CardType           string           `json:"card_type" db:"card_type"`
	ExpireDate         pgtype.Date      `json:"expire_date" db:"expire_date"`
	Cvv                string           `json:"cvv" db:"cvv"`
	CardProvider       string           `json:"card_provider" db:"card_provider"`
	Status             string           `json:"status" db:"status"`
	CreditLimit        int32            `json:"credit_limit" db:"credit_limit"`
	OutstandingBalance int32            `json:"outstanding_balance" db:"outstanding_balance"`
	RewardPoints       int32            `json:"reward_points" db:"reward_points"`
	CreatedAt          pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt          pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type RedeemRewardPointsRow struct {
	CardID             int32            `json:"card_id" db:"card_id"`
	UserID             int32            `json:"user_id" db:"user_id"`
	CardNumber         string           `json:"card_number" db:"card_number"`
	CardType           string           `json:"card_type" db:"card_type"`
	ExpireDate         pgtype.Date      `json:"expire_date" db:"expire_date"`
	Cvv                string           `json:"cvv" db:"cvv"`
	CardProvider       string           `json:"card_provider" db:"card_provider"`
	Status             string           `json:"status" db:"status"`
	CreditLimit        int32            `json:"credit_limit" db:"credit_limit"`
	OutstandingBalance int32            `json:"outstanding_balance" db:"outstanding_balance"`
	RewardPoints       int32            `json:"reward_points" db:"reward_points"`
	CreatedAt          pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt          pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type GetAuthTransactionsByCardNumberRow struct {
	AuthID         int32            `json:"auth_id" db:"auth_id"`
	TxnID          string           `json:"txn_id" db:"txn_id"`
	CardNumber     string           `json:"card_number" db:"card_number"`
	MerchantID     int32            `json:"merchant_id" db:"merchant_id"`
	Amount         int64            `json:"amount" db:"amount"`
	Currency       string           `json:"currency" db:"currency"`
	Mcc            string           `json:"mcc" db:"mcc"`
	PosEntryMode   string           `json:"pos_entry_mode" db:"pos_entry_mode"`
	Status         string           `json:"status" db:"status"`
	IdempotencyKey string           `json:"idempotency_key" db:"idempotency_key"`
	RiskScore      int32            `json:"risk_score" db:"risk_score"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	TotalCount     int64            `json:"total_count" db:"total_count"`
}

type GetCardPaymentsByCardNumberRow struct {
	PaymentID      int32            `json:"payment_id" db:"payment_id"`
	PaymentUuid    string           `json:"payment_uuid" db:"payment_uuid"`
	CardNumber     string           `json:"card_number" db:"card_number"`
	BillingID      *int32           `json:"billing_id" db:"billing_id"`
	Amount         int64            `json:"amount" db:"amount"`
	PaymentChannel string           `json:"payment_channel" db:"payment_channel"`
	ReferenceID    string           `json:"reference_id" db:"reference_id"`
	Status         string           `json:"status" db:"status"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	TotalCount     int64            `json:"total_count" db:"total_count"`
}

type GetRedeemableRewardIdsRow struct {
	RewardID     int32 `json:"reward_id" db:"reward_id"`
	PointsEarned int32 `json:"points_earned" db:"points_earned"`
}

type GetRolesRow struct {
	RoleID     int32            `json:"role_id" db:"role_id"`
	RoleName   string           `json:"role_name" db:"role_name"`
	CreatedAt  pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt  pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	TotalCount int64            `json:"total_count" db:"total_count"`
}

type GetActiveRolesRow struct {
	RoleID     int32            `json:"role_id" db:"role_id"`
	RoleName   string           `json:"role_name" db:"role_name"`
	CreatedAt  pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt  pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt  pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
	TotalCount int64            `json:"total_count" db:"total_count"`
}

type GetTrashedRolesRow struct {
	RoleID     int32            `json:"role_id" db:"role_id"`
	RoleName   string           `json:"role_name" db:"role_name"`
	CreatedAt  pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt  pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt  pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
	TotalCount int64            `json:"total_count" db:"total_count"`
}
