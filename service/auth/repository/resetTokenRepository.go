package repository

import (
	"context"
	"time"

	database "github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database"
	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/auth/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/jmoiron/sqlx"
)

// SQL source of truth (formerly database/query/reset_token.sql).
const (
	createResetToken = `INSERT INTO
    reset_tokens (user_id, token, expiry_date)
VALUES ($1, $2, $3)
RETURNING
    user_id,
    token,
    expiry_date;`

	deleteResetToken = `DELETE FROM reset_tokens WHERE user_id = $1;`

	getResetToken = `SELECT user_id, token, expiry_date
FROM reset_tokens
WHERE
    token = $1;`
)

// resetTokenRepository is a struct that implements the ResetTokenRepository interface
type resetTokenRepository struct {
	db *sqlx.DB
}

// NewResetTokenRepository creates a new instance of resetTokenRepository.
func NewResetTokenRepository(db *sqlx.DB) *resetTokenRepository {
	return &resetTokenRepository{
		db: db,
	}
}

// FindByToken retrieves a reset token record by token string.
func (r *resetTokenRepository) FindByToken(ctx context.Context, code string) (*db.GetResetTokenRow, error) {
	var res db.GetResetTokenRow
	err := r.db.GetContext(ctx, &res, getResetToken, code)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &res, nil
}

// CreateResetToken inserts a new reset token into the database.
func (r *resetTokenRepository) CreateResetToken(ctx context.Context, req *requests.CreateResetTokenRequest) (*db.CreateResetTokenRow, error) {
	expiryDate, err := time.Parse("2006-01-02 15:04:05", req.ExpiredAt)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	var res db.CreateResetTokenRow
	err = r.db.GetContext(ctx, &res, createResetToken,
		int32(req.UserID),
		req.ResetToken,
		expiryDate,
	)
	if err != nil {
		if database.IsUniqueViolation(err) {
			return nil, sharedErrors.NewConflictError("reset token already exists").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("create reset token").WithInternal(err)
	}
	return &res, nil
}

// DeleteResetToken removes the reset token associated with the given user ID.
func (r *resetTokenRepository) DeleteResetToken(ctx context.Context, user_id int32) error {
	_, err := r.db.ExecContext(ctx, deleteResetToken, user_id)
	if err != nil {
		return sharedErrors.ErrFailed("delete reset token").WithInternal(err)
	}
	return nil
}
