package repository

import (
	"context"

	database "github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database"
	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const createUser = `INSERT INTO
    users (
        firstname,
        lastname,
        email,
        password,
        verification_code,
        is_verified,
        created_at,
        updated_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        current_timestamp,
        current_timestamp
    )
RETURNING
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at`

const updateUser = `UPDATE users
SET
    firstname = $2,
    lastname = $3,
    email = $4,
    password = $5,
    updated_at = current_timestamp
WHERE
    user_id = $1
    AND deleted_at IS NULL
RETURNING
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at`

const updateUserIsVerified = `UPDATE users
SET
    is_verified = $2,
    updated_at = current_timestamp
WHERE
    user_id = $1
    AND deleted_at IS NULL
RETURNING
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at`

const updateUserPassword = `UPDATE users
SET
    password = $2,
    updated_at = current_timestamp
WHERE
    user_id = $1
    AND deleted_at IS NULL
RETURNING
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at`

const trashUser = `UPDATE users
SET
    deleted_at = current_timestamp,
    updated_at = current_timestamp
WHERE
    user_id = $1
    AND deleted_at IS NULL
RETURNING
    user_id,
    firstname,
    lastname,
    email,
    created_at,
    updated_at,
    deleted_at`

const restoreUser = `UPDATE users
SET
    deleted_at = NULL,
    updated_at = current_timestamp
WHERE
    user_id = $1
    AND deleted_at IS NOT NULL
RETURNING
    user_id,
    firstname,
    lastname,
    email,
    created_at,
    updated_at,
    deleted_at`

const deleteUserPermanently = `DELETE FROM users WHERE user_id = $1 AND deleted_at IS NOT NULL`

const restoreAllUsers = `UPDATE users
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL`

const deleteAllPermanentUsers = `DELETE FROM users WHERE deleted_at IS NOT NULL`

// userCommandRepository is a struct that implements the UserCommandRepository interface
type userCommandRepository struct {
	db *sqlx.DB
}

// NewUserCommandRepository creates a new instance of userCommandRepository.
func NewUserCommandRepository(db *sqlx.DB) UserCommandRepository {
	return &userCommandRepository{
		db: db,
	}
}

// CreateUser inserts a new user record into the database.
func (r *userCommandRepository) CreateUser(ctx context.Context, request *requests.CreateUserRequest) (*db.CreateUserRow, error) {
	verified := false
	verifyCode := uuid.New().String()

	var user db.CreateUserRow

	err := r.db.GetContext(ctx, &user, createUser,
		request.FirstName,
		request.LastName,
		request.Email,
		request.Password,
		verifyCode,
		&verified,
	)

	if err != nil {
		if database.IsUniqueViolation(err) {
			return nil, sharedErrors.NewConflictError("email already exists").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("create user").WithInternal(err)
	}

	return &user, nil
}

// UpdateUser updates an existing user record in the database.
func (r *userCommandRepository) UpdateUser(ctx context.Context, request *requests.UpdateUserRequest) (*db.UpdateUserRow, error) {
	var res db.UpdateUserRow

	err := r.db.GetContext(ctx, &res, updateUser,
		int32(*request.UserID),
		request.FirstName,
		request.LastName,
		request.Email,
		request.Password,
	)

	if err != nil {
		if database.IsUniqueViolation(err) {
			return nil, sharedErrors.NewConflictError("email already exists").WithInternal(err)
		}
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "user", "update user")
	}

	return &res, nil
}

// TrashedUser soft-deletes a user by marking it as trashed.
func (r *userCommandRepository) TrashedUser(ctx context.Context, user_id int) (*db.TrashUserRow, error) {
	var res db.TrashUserRow

	err := r.db.GetContext(ctx, &res, trashUser, int32(user_id))

	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "user", "trash user")
	}

	return &res, nil
}

// RestoreUser restores a soft-deleted (trashed) user.
func (r *userCommandRepository) RestoreUser(ctx context.Context, user_id int) (*db.RestoreUserRow, error) {
	var res db.RestoreUserRow

	err := r.db.GetContext(ctx, &res, restoreUser, int32(user_id))

	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "user", "restore user")
	}

	return &res, nil
}

// DeleteUserPermanent permanently deletes a user from the database.
func (r *userCommandRepository) DeleteUserPermanent(ctx context.Context, user_id int) (bool, error) {
	result, err := r.db.ExecContext(ctx, deleteUserPermanently, int32(user_id))

	if err != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(err, "user", "delete user permanently")
	}

	if count, err := result.RowsAffected(); err == nil && count == 0 {
		return false, sharedErrors.ErrNotFound.WithMessage("user not found")
	}

	return true, nil
}

// RestoreAllUser restores all soft-deleted users in the database.
func (r *userCommandRepository) RestoreAllUser(ctx context.Context) (bool, error) {
	_, err := r.db.ExecContext(ctx, restoreAllUsers)

	if err != nil {
		return false, sharedErrors.ErrFailed("restore all users").WithInternal(err)
	}

	return true, nil
}

// DeleteAllUserPermanent permanently deletes all trashed users from the database.
func (r *userCommandRepository) DeleteAllUserPermanent(ctx context.Context) (bool, error) {
	_, err := r.db.ExecContext(ctx, deleteAllPermanentUsers)

	if err != nil {
		return false, sharedErrors.ErrFailed("delete all users permanently").WithInternal(err)
	}
	return true, nil
}

func (r *userCommandRepository) UpdateIsVerified(ctx context.Context, userID int, isVerified bool) (*db.UpdateUserIsVerifiedRow, error) {
	var res db.UpdateUserIsVerifiedRow

	err := r.db.GetContext(ctx, &res, updateUserIsVerified, int32(userID), &isVerified)

	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "user", "update user")
	}

	return &res, nil
}

func (r *userCommandRepository) UpdatePassword(ctx context.Context, userID int, password string) (*db.UpdateUserPasswordRow, error) {
	var res db.UpdateUserPasswordRow

	err := r.db.GetContext(ctx, &res, updateUserPassword, int32(userID), password)

	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "user", "update user")
	}

	return &res, nil
}
