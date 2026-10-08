package repository

import (
	"context"

	database "github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database"
	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/role/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/jmoiron/sqlx"
)

// SQL source of truth (formerly database/query/role.sql and user_role.sql).
const (
	createRole = `INSERT INTO
    roles (
        role_name,
        created_at,
        updated_at
    )
VALUES (
        $1,
        current_timestamp,
        current_timestamp
    )
RETURNING
    role_id,
    role_name,
    created_at,
    updated_at,
    deleted_at;`

	updateRole = `UPDATE roles
SET
    role_name = $2,
    updated_at = current_timestamp
WHERE
    role_id = $1
RETURNING
    role_id,
    role_name,
    created_at,
    updated_at,
    deleted_at;`

	trashRole = `UPDATE roles
SET
    deleted_at = current_timestamp
WHERE
    role_id = $1
RETURNING
    *;`

	restoreRole = `UPDATE roles
SET
    deleted_at = NULL
WHERE
    role_id = $1
RETURNING
    *;`

	deletePermanentRole = `DELETE FROM roles
WHERE
    role_id = $1
    AND deleted_at IS NOT NULL
RETURNING
    *;`

	restoreAllRoles = `UPDATE roles
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL;`

	deleteAllPermanentRoles = `DELETE FROM roles WHERE deleted_at IS NOT NULL;`

	assignRoleToUser = `INSERT INTO user_roles (
    user_id,
    role_id,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    current_timestamp,
    current_timestamp
) RETURNING
    user_role_id,
    user_id,
    role_id,
    created_at,
    updated_at,
    deleted_at;`

	removeRoleFromUser = `DELETE FROM user_roles
WHERE
    user_id = $1
    AND role_id = $2;`
)

// roleCommandRepository is a struct that implements the RoleCommandRepository interface
type roleCommandRepository struct {
	db *sqlx.DB
}

// NewRoleCommandRepository creates a new RoleCommandRepository instance with the provided
// database handle. This repository is responsible for executing command operations
// related to role records in the database.
//
// Parameters:
//   - db: A pointer to the sqlx.DB object for executing database queries.
//
// Returns:
//   - A newly created roleCommandRepository instance.
func NewRoleCommandRepository(db *sqlx.DB) RoleCommandRepository {
	return &roleCommandRepository{
		db: db,
	}
}

func (r *roleCommandRepository) CreateRole(ctx context.Context, req *requests.CreateRoleRequest) (*db.Role, error) {
	var res db.Role
	err := r.db.GetContext(ctx, &res, createRole, req.Name)

	if err != nil {
		if database.IsUniqueViolation(err) {
			return nil, sharedErrors.NewConflictError("role name already exists").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("create role").WithInternal(err)
	}

	return &res, nil
}

func (r *roleCommandRepository) UpdateRole(ctx context.Context, req *requests.UpdateRoleRequest) (*db.Role, error) {
	var res db.Role
	err := r.db.GetContext(ctx, &res, updateRole,
		int32(*req.ID),
		req.Name,
	)

	if err != nil {
		if database.IsUniqueViolation(err) {
			return nil, sharedErrors.NewConflictError("role name already exists").WithInternal(err)
		}
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "role", "update role")
	}

	return &res, nil
}

func (r *roleCommandRepository) TrashedRole(ctx context.Context, id int) (*db.Role, error) {
	var res db.Role
	err := r.db.GetContext(ctx, &res, trashRole, int32(id))
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "role", "trash role")
	}
	return &res, nil
}

func (r *roleCommandRepository) RestoreRole(ctx context.Context, id int) (*db.Role, error) {
	var res db.Role
	err := r.db.GetContext(ctx, &res, restoreRole, int32(id))
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "role", "restore role")
	}
	return &res, nil
}

func (r *roleCommandRepository) DeleteRolePermanent(ctx context.Context, role_id int) (bool, error) {
	_, err := r.db.ExecContext(ctx, deletePermanentRole, int32(role_id))
	if err != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(err, "role", "delete role permanently")
	}
	return true, nil
}

func (r *roleCommandRepository) RestoreAllRole(ctx context.Context) (bool, error) {
	_, err := r.db.ExecContext(ctx, restoreAllRoles)

	if err != nil {
		return false, sharedErrors.ErrFailed("restore all roles").WithInternal(err)
	}

	return true, nil
}

func (r *roleCommandRepository) DeleteAllRolePermanent(ctx context.Context) (bool, error) {
	_, err := r.db.ExecContext(ctx, deleteAllPermanentRoles)

	if err != nil {
		return false, sharedErrors.ErrFailed("delete all roles").WithInternal(err)
	}

	return true, nil
}

func (r *roleCommandRepository) CreateUserRole(ctx context.Context, userID, roleID int) (*db.Role, error) {
	_, err := r.db.ExecContext(ctx, assignRoleToUser,
		int32(userID),
		int32(roleID),
	)

	if err != nil {
		return nil, sharedErrors.ErrFailed("create user role").WithInternal(err)
	}

	// This method returns *db.UserRole from the insert, but we need *db.Role or similar.
	// Since the interface says *db.Role, I might need to fetch the role or just return the role info.
	// Actually, AssignRoleToUser returns *db.UserRole which doesn't contain RoleName.
	// I'll fetch the role info to satisfy the contract.
	var role db.Role
	err = r.db.GetContext(ctx, &role, getRole, int32(roleID))
	if err != nil {
		return nil, sharedErrors.ErrFailed("get role for user role").WithInternal(err)
	}

	return &role, nil
}

func (r *roleCommandRepository) DeleteUserRole(ctx context.Context, userID, roleID int) (bool, error) {
	_, err := r.db.ExecContext(ctx, removeRoleFromUser,
		int32(userID),
		int32(roleID),
	)

	if err != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(err, "user role", "delete user role")
	}

	return true, nil
}
