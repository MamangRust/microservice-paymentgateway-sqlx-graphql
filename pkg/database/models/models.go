// Package models defines the cross-service domain shapes returned by the
// pkg/adapter gRPC adapters. Consuming services depend on these structs (not
// on generated pb types) so the proto surface stays behind the adapter.
package models

import "time"

type User struct {
	UserID    int32      `json:"user_id"`
	Firstname string     `json:"firstname"`
	Lastname  string     `json:"lastname"`
	Email     string     `json:"email"`
	Password  string     `json:"password"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type Card struct {
	CardID       int32      `json:"card_id"`
	UserID       int32      `json:"user_id"`
	CardNumber   string     `json:"card_number"`
	CardType     string     `json:"card_type"`
	ExpireDate   *time.Time `json:"expire_date"`
	Cvv          string     `json:"cvv"`
	CardProvider string     `json:"card_provider"`
	Email        string     `json:"email"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

type Merchant struct {
	MerchantID int32      `json:"merchant_id"`
	UserID     int32      `json:"user_id"`
	Name       string     `json:"name"`
	ApiKey     string     `json:"api_key"`
	Status     string     `json:"status"`
	CreatedAt  *time.Time `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at"`
}

type Saldo struct {
	SaldoID        int32      `json:"saldo_id"`
	CardNumber     string     `json:"card_number"`
	TotalBalance   int64      `json:"total_balance"`
	WithdrawAmount *int64     `json:"withdraw_amount"`
	WithdrawTime   *time.Time `json:"withdraw_time"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
}

// SaldoMutationResult is the balance snapshot returned by saldo mutation
// calls (update/debit/credit/withdraw) through the SaldoAdapter.
type SaldoMutationResult struct {
	SaldoID      int32  `json:"saldo_id"`
	CardNumber   string `json:"card_number"`
	TotalBalance int64  `json:"total_balance"`
}

type Role struct {
	RoleID    int32      `json:"role_id"`
	RoleName  string     `json:"role_name"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type UserRole struct {
	UserRoleID int32 `json:"user_role_id"`
	UserID     int32 `json:"user_id"`
	RoleID     int32 `json:"role_id"`
}

// Transaction is the cross-service transaction shape returned by the
// TransactionAdapter. It mirrors the fields the merchant service needs to render
// its merchant-transaction listings without reading the transaction table directly.
type Transaction struct {
	TransactionID   int        `json:"transaction_id"`
	TransactionNo   string     `json:"transaction_no"`
	CardNumber      string     `json:"card_number"`
	Amount          int64      `json:"amount"`
	PaymentMethod   string     `json:"payment_method"`
	MerchantID      int        `json:"merchant_id"`
	MerchantName    string     `json:"merchant_name"`
	TransactionTime time.Time  `json:"transaction_time"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at"`
}
