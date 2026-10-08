// Package db holds the hand-written data models for the card service.
//
// This file replaces the sqlc-generated models: struct shapes, field types
// and json tags are unchanged from the generated version, with sqlx `db`
// tags added for column mapping.
package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type BillingCycle struct {
	BillingID  int32            `json:"billing_id" db:"billing_id"`
	CardNumber string           `json:"card_number" db:"card_number"`
	CycleStart time.Time        `json:"cycle_start" db:"cycle_start"`
	CycleEnd   time.Time        `json:"cycle_end" db:"cycle_end"`
	AmountDue  int32            `json:"amount_due" db:"amount_due"`
	DueDate    time.Time        `json:"due_date" db:"due_date"`
	Status     string           `json:"status" db:"status"`
	CreatedAt  pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt  pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type Card struct {
	CardID             int32            `json:"card_id" db:"card_id"`
	UserID             int32            `json:"user_id" db:"user_id"`
	CardNumber         string           `json:"card_number" db:"card_number"`
	CardType           string           `json:"card_type" db:"card_type"`
	ExpireDate         pgtype.Date      `json:"expire_date" db:"expire_date"`
	Cvv                string           `json:"cvv" db:"cvv"`
	CardProvider       string           `json:"card_provider" db:"card_provider"`
	CreatedAt          pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt          pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt          pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
	Status             string           `json:"status" db:"status"`
	CreditLimit        int32            `json:"credit_limit" db:"credit_limit"`
	OutstandingBalance int32            `json:"outstanding_balance" db:"outstanding_balance"`
	RewardPoints       int32            `json:"reward_points" db:"reward_points"`
}

type CardAuthTransaction struct {
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
}

type CardPayment struct {
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
}

type CardReward struct {
	RewardID     int32            `json:"reward_id" db:"reward_id"`
	CardNumber   string           `json:"card_number" db:"card_number"`
	TxnID        string           `json:"txn_id" db:"txn_id"`
	Amount       int64            `json:"amount" db:"amount"`
	Mcc          string           `json:"mcc" db:"mcc"`
	PointsEarned int32            `json:"points_earned" db:"points_earned"`
	ExpiresAt    time.Time        `json:"expires_at" db:"expires_at"`
	Redeemed     bool             `json:"redeemed" db:"redeemed"`
	CreatedAt    pgtype.Timestamp `json:"created_at" db:"created_at"`
}

type Role struct {
	RoleID    int32            `json:"role_id" db:"role_id"`
	RoleName  string           `json:"role_name" db:"role_name"`
	CreatedAt pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}

type UserRole struct {
	UserRoleID int32            `json:"user_role_id" db:"user_role_id"`
	UserID     int32            `json:"user_id" db:"user_id"`
	RoleID     int32            `json:"role_id" db:"role_id"`
	CreatedAt  pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt  pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt  pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}

// CardStatusTotal is the aggregate row type for status counts over cards.
type CardStatusTotal struct {
	Status     string `json:"status" db:"status"`
	TotalCount int64  `json:"total_count" db:"total_count"`
}

// CardRewardPointsTotal is the aggregate row type for SUM(points_earned).
type CardRewardPointsTotal struct {
	Points int64 `json:"points" db:"points"`
}

// CardAuthTransactionCount is the aggregate row type for COUNT(*) over
// card auth transactions.
type CardAuthTransactionCount struct {
	TotalCount int32 `json:"total_count" db:"total_count"`
}
