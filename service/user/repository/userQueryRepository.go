package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/jmoiron/sqlx"
)

const getUsersWithPagination = `SELECT
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM users
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR firstname ILIKE '%' || $1 || '%'
        OR lastname ILIKE '%' || $1 || '%'
        OR email ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET $3`

const getActiveUsersWithPagination = `SELECT
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM users
WHERE
    deleted_at IS NOT NULL
    AND (
        $1::TEXT IS NULL
        OR firstname ILIKE '%' || $1 || '%'
        OR lastname ILIKE '%' || $1 || '%'
        OR email ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET $3`

const getTrashedUsersWithPagination = `SELECT
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM users
WHERE
    deleted_at IS NOT NULL
    AND (
        $1::TEXT IS NULL
        OR firstname ILIKE '%' || $1 || '%'
        OR lastname ILIKE '%' || $1 || '%'
        OR email ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET $3`

const getUserByID = `SELECT
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at
FROM users
WHERE
    user_id = $1
    AND deleted_at IS NULL`

const getUserByEmail = `SELECT user_id, firstname, lastname, email, password, created_at, updated_at
FROM users
WHERE
    email = $1
    AND deleted_at IS NULL`

const getUserByVerificationCode = `SELECT
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at
FROM users
WHERE
    verification_code = $1`

// userQueryRepository implements UserQueryRepository.
type userQueryRepository struct {
	db *sqlx.DB
}

// NewUserQueryRepository creates a new UserQueryRepository.
func NewUserQueryRepository(db *sqlx.DB) UserQueryRepository {
	return &userQueryRepository{
		db: db,
	}
}

func (r *userQueryRepository) FindAllUsers(ctx context.Context, req *requests.FindAllUsers) ([]*db.GetUsersWithPaginationRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var rows []*db.GetUsersWithPaginationRow

	err := r.db.SelectContext(ctx, &rows, getUsersWithPagination, req.Search, int32(req.PageSize), int32(offset))

	if err != nil {
		return nil, sharedErrors.ErrFailed("find all users").WithInternal(err)
	}

	return rows, nil
}

func (r *userQueryRepository) FindById(ctx context.Context, user_id int) (*db.GetUserByIDRow, error) {
	var user db.GetUserByIDRow

	err := r.db.GetContext(ctx, &user, getUserByID, int32(user_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFound.WithMessage("user not found").WithInternal(err)
		}

		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &user, nil
}

func (r *userQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllUsers) ([]*db.GetActiveUsersWithPaginationRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var rows []*db.GetActiveUsersWithPaginationRow

	err := r.db.SelectContext(ctx, &rows, getActiveUsersWithPagination, req.Search, int32(req.PageSize), int32(offset))

	if err != nil {
		return nil, sharedErrors.ErrFailed("find active users").WithInternal(err)
	}

	return rows, nil
}

func (r *userQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllUsers) ([]*db.GetTrashedUsersWithPaginationRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var rows []*db.GetTrashedUsersWithPaginationRow

	err := r.db.SelectContext(ctx, &rows, getTrashedUsersWithPagination, req.Search, int32(req.PageSize), int32(offset))

	if err != nil {
		return nil, sharedErrors.ErrFailed("find trashed users").WithInternal(err)
	}

	return rows, nil
}

func (r *userQueryRepository) FindByEmail(ctx context.Context, email string) (*db.GetUserByEmailRow, error) {
	var user db.GetUserByEmailRow

	err := r.db.GetContext(ctx, &user, getUserByEmail, email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFound.WithMessage("user not found").WithInternal(err)
		}

		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &user, nil
}

func (r *userQueryRepository) FindByVerificationCode(ctx context.Context, code string) (*db.GetUserByVerificationCodeRow, error) {
	var user db.GetUserByVerificationCodeRow

	err := r.db.GetContext(ctx, &user, getUserByVerificationCode, code)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFound.WithMessage("user not found").WithInternal(err)
		}

		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &user, nil
}
