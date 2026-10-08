// Row structs extracted from the former sqlc *.sql.go files (migrated to sqlx).
package db

import (
	"time"
)

type CreateResetTokenRow struct {
	UserID     int32     `json:"user_id" db:"user_id"`
	Token      string    `json:"token" db:"token"`
	ExpiryDate time.Time `json:"expiry_date" db:"expiry_date"`
}

type GetResetTokenRow struct {
	UserID     int32     `json:"user_id" db:"user_id"`
	Token      string    `json:"token" db:"token"`
	ExpiryDate time.Time `json:"expiry_date" db:"expiry_date"`
}
