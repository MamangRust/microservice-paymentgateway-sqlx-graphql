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

// SQL source of truth (formerly database/query/refresh_token.sql).
const (
	createRefreshToken = `INSERT INTO refresh_tokens (user_id, token, expiration, created_at, updated_at)
VALUES ($1, $2, $3, current_timestamp, current_timestamp)
RETURNING refresh_token_id, user_id, token, expiration, created_at, updated_at, deleted_at;`

	findRefreshTokenByToken = `SELECT refresh_token_id, user_id, token, expiration, created_at, updated_at, deleted_at
FROM refresh_tokens
WHERE token = $1 AND deleted_at IS NULL;`

	findRefreshTokenByUserId = `SELECT
    refresh_token_id,
    user_id,
    token,
    expiration,
    created_at,
    updated_at,
    deleted_at
FROM
    refresh_tokens
WHERE
    user_id = $1 AND deleted_at IS NULL
ORDER BY
    created_at DESC
LIMIT 1;`

	updateRefreshTokenByUserId = `UPDATE refresh_tokens
SET token = $2, expiration = $3, updated_at = current_timestamp
WHERE user_id = $1 AND deleted_at IS NULL
RETURNING *;`

	deleteRefreshToken = `DELETE FROM refresh_tokens
WHERE token = $1;`

	deleteRefreshTokenByUserId = `DELETE FROM refresh_tokens
WHERE user_id = $1;`
)

// refreshTokenRepository is a struct that implements the RefreshTokenRepository interface
type refreshTokenRepository struct {
	db *sqlx.DB
}

// NewRefreshTokenRepository creates a new RefreshTokenRepository instance
func NewRefreshTokenRepository(db *sqlx.DB) *refreshTokenRepository {
	return &refreshTokenRepository{
		db: db,
	}
}

// FindByToken retrieves a refresh token record by the token string.
func (r *refreshTokenRepository) FindByToken(ctx context.Context, token string) (*db.RefreshToken, error) {
	var res db.RefreshToken
	err := r.db.GetContext(ctx, &res, findRefreshTokenByToken, token)

	if err != nil {
		return nil, sharedErrors.ErrTokenNotFound.WithInternal(err)
	}

	return &res, nil
}

// FindByUserId retrieves a refresh token record by the associated user ID.
func (r *refreshTokenRepository) FindByUserId(ctx context.Context, user_id int) (*db.RefreshToken, error) {
	var res db.RefreshToken
	err := r.db.GetContext(ctx, &res, findRefreshTokenByUserId, int32(user_id))

	if err != nil {
		return nil, sharedErrors.ErrFindByUserID.WithInternal(err)
	}

	return &res, nil
}

// CreateRefreshToken inserts a new refresh token record into the database.
func (r *refreshTokenRepository) CreateRefreshToken(ctx context.Context, req *requests.CreateRefreshToken) (*db.RefreshToken, error) {
	layout := "2006-01-02 15:04:05"
	expirationTime, err := time.Parse(layout, req.ExpiresAt)
	if err != nil {
		return nil, sharedErrors.ErrParseDate.WithInternal(err)
	}

	var res db.RefreshToken
	err = r.db.GetContext(ctx, &res, createRefreshToken,
		int32(req.UserId),
		req.Token,
		expirationTime,
	)

	if err != nil {
		if database.IsUniqueViolation(err) {
			return nil, sharedErrors.NewConflictError("refresh token already exists").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("create refresh token").WithInternal(err)
	}

	return &res, nil
}

// UpdateRefreshToken updates an existing refresh token record.
func (r *refreshTokenRepository) UpdateRefreshToken(ctx context.Context, req *requests.UpdateRefreshToken) (*db.RefreshToken, error) {
	layout := "2006-01-02 15:04:05"
	expirationTime, err := time.Parse(layout, req.ExpiresAt)
	if err != nil {
		return nil, sharedErrors.ErrParseDate.WithInternal(err)
	}

	var res db.RefreshToken
	err = r.db.GetContext(ctx, &res, updateRefreshTokenByUserId,
		int32(req.UserId),
		req.Token,
		expirationTime,
	)
	if err != nil {
		if database.IsUniqueViolation(err) {
			return nil, sharedErrors.NewConflictError("refresh token already exists").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("update refresh token").WithInternal(err)
	}

	return &res, nil
}

// DeleteRefreshToken removes a refresh token by its token string.
func (r *refreshTokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, deleteRefreshToken, token)

	if err != nil {
		return sharedErrors.ErrFailed("delete refresh token").WithInternal(err)
	}

	return nil
}

// DeleteRefreshTokenByUserId removes a refresh token by the associated user ID.
func (r *refreshTokenRepository) DeleteRefreshTokenByUserId(ctx context.Context, user_id int) error {
	_, err := r.db.ExecContext(ctx, deleteRefreshTokenByUserId, int32(user_id))

	if err != nil {
		return sharedErrors.ErrFailed("delete refresh token by user ID").WithInternal(err)
	}

	return nil
}
