package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type CreateUserParams struct {
	Firstname        string `db:"firstname" json:"firstname"`
	Lastname         string `db:"lastname" json:"lastname"`
	Email            string `db:"email" json:"email"`
	Password         string `db:"password" json:"password"`
	VerificationCode string `db:"verification_code" json:"verification_code"`
	IsVerified       *bool  `db:"is_verified" json:"is_verified"`
}

type CreateUserRow struct {
	UserID    int32            `db:"user_id" json:"user_id"`
	Firstname string           `db:"firstname" json:"firstname"`
	Lastname  string           `db:"lastname" json:"lastname"`
	Email     string           `db:"email" json:"email"`
	Password  string           `db:"password" json:"password"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetActiveUsersWithPaginationParams struct {
	Column1 string `db:"column_1" json:"column_1"`
	Limit   int32  `db:"limit" json:"limit"`
	Offset  int32  `db:"offset" json:"offset"`
}

type GetActiveUsersWithPaginationRow struct {
	UserID     int32            `db:"user_id" json:"user_id"`
	Firstname  string           `db:"firstname" json:"firstname"`
	Lastname   string           `db:"lastname" json:"lastname"`
	Email      string           `db:"email" json:"email"`
	Password   string           `db:"password" json:"password"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type GetTrashedUsersWithPaginationParams struct {
	Column1 string `db:"column_1" json:"column_1"`
	Limit   int32  `db:"limit" json:"limit"`
	Offset  int32  `db:"offset" json:"offset"`
}

type GetTrashedUsersWithPaginationRow struct {
	UserID     int32            `db:"user_id" json:"user_id"`
	Firstname  string           `db:"firstname" json:"firstname"`
	Lastname   string           `db:"lastname" json:"lastname"`
	Email      string           `db:"email" json:"email"`
	Password   string           `db:"password" json:"password"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type GetUserByEmailRow struct {
	UserID    int32            `db:"user_id" json:"user_id"`
	Firstname string           `db:"firstname" json:"firstname"`
	Lastname  string           `db:"lastname" json:"lastname"`
	Email     string           `db:"email" json:"email"`
	Password  string           `db:"password" json:"password"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetUserByIDRow struct {
	UserID    int32            `db:"user_id" json:"user_id"`
	Firstname string           `db:"firstname" json:"firstname"`
	Lastname  string           `db:"lastname" json:"lastname"`
	Email     string           `db:"email" json:"email"`
	Password  string           `db:"password" json:"password"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetUserByVerificationCodeRow struct {
	UserID    int32            `db:"user_id" json:"user_id"`
	Firstname string           `db:"firstname" json:"firstname"`
	Lastname  string           `db:"lastname" json:"lastname"`
	Email     string           `db:"email" json:"email"`
	Password  string           `db:"password" json:"password"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetUsersWithPaginationParams struct {
	Column1 string `db:"column_1" json:"column_1"`
	Limit   int32  `db:"limit" json:"limit"`
	Offset  int32  `db:"offset" json:"offset"`
}

type GetUsersWithPaginationRow struct {
	UserID     int32            `db:"user_id" json:"user_id"`
	Firstname  string           `db:"firstname" json:"firstname"`
	Lastname   string           `db:"lastname" json:"lastname"`
	Email      string           `db:"email" json:"email"`
	Password   string           `db:"password" json:"password"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type RestoreUserRow struct {
	UserID    int32            `db:"user_id" json:"user_id"`
	Firstname string           `db:"firstname" json:"firstname"`
	Lastname  string           `db:"lastname" json:"lastname"`
	Email     string           `db:"email" json:"email"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}

type TrashUserRow struct {
	UserID    int32            `db:"user_id" json:"user_id"`
	Firstname string           `db:"firstname" json:"firstname"`
	Lastname  string           `db:"lastname" json:"lastname"`
	Email     string           `db:"email" json:"email"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}

type UpdateUserParams struct {
	UserID    int32  `db:"user_id" json:"user_id"`
	Firstname string `db:"firstname" json:"firstname"`
	Lastname  string `db:"lastname" json:"lastname"`
	Email     string `db:"email" json:"email"`
	Password  string `db:"password" json:"password"`
}

type UpdateUserRow struct {
	UserID    int32            `db:"user_id" json:"user_id"`
	Firstname string           `db:"firstname" json:"firstname"`
	Lastname  string           `db:"lastname" json:"lastname"`
	Email     string           `db:"email" json:"email"`
	Password  string           `db:"password" json:"password"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type UpdateUserIsVerifiedParams struct {
	UserID     int32 `db:"user_id" json:"user_id"`
	IsVerified *bool `db:"is_verified" json:"is_verified"`
}

type UpdateUserIsVerifiedRow struct {
	UserID    int32            `db:"user_id" json:"user_id"`
	Firstname string           `db:"firstname" json:"firstname"`
	Lastname  string           `db:"lastname" json:"lastname"`
	Email     string           `db:"email" json:"email"`
	Password  string           `db:"password" json:"password"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type UpdateUserPasswordParams struct {
	UserID   int32  `db:"user_id" json:"user_id"`
	Password string `db:"password" json:"password"`
}

type UpdateUserPasswordRow struct {
	UserID    int32            `db:"user_id" json:"user_id"`
	Firstname string           `db:"firstname" json:"firstname"`
	Lastname  string           `db:"lastname" json:"lastname"`
	Email     string           `db:"email" json:"email"`
	Password  string           `db:"password" json:"password"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetUserByEmailWithPasswordRow struct {
	UserID   int32  `db:"user_id" json:"user_id"`
	Email    string `db:"email" json:"email"`
	Password string `db:"password" json:"password"`
}

type GetUserByEmailAndVerifiedRow struct {
	UserID    int32            `db:"user_id" json:"user_id"`
	Firstname string           `db:"firstname" json:"firstname"`
	Lastname  string           `db:"lastname" json:"lastname"`
	Email     string           `db:"email" json:"email"`
	Password  string           `db:"password" json:"password"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type CreateResetTokenParams struct {
	UserID     int32     `db:"user_id" json:"user_id"`
	Token      string    `db:"token" json:"token"`
	ExpiryDate time.Time `db:"expiry_date" json:"expiry_date"`
}

type CreateResetTokenRow struct {
	UserID     int32     `db:"user_id" json:"user_id"`
	Token      string    `db:"token" json:"token"`
	ExpiryDate time.Time `db:"expiry_date" json:"expiry_date"`
}

type GetResetTokenRow struct {
	UserID     int32     `db:"user_id" json:"user_id"`
	Token      string    `db:"token" json:"token"`
	ExpiryDate time.Time `db:"expiry_date" json:"expiry_date"`
}

type AssignRoleToUserParams struct {
	UserID int32 `db:"user_id" json:"user_id"`
	RoleID int32 `db:"role_id" json:"role_id"`
}

type GetTrashedUserRolesRow struct {
	UserRoleID int32            `db:"user_role_id" json:"user_role_id"`
	UserID     int32            `db:"user_id" json:"user_id"`
	RoleID     int32            `db:"role_id" json:"role_id"`
	RoleName   string           `db:"role_name" json:"role_name"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}

type RemoveRoleFromUserParams struct {
	UserID int32 `db:"user_id" json:"user_id"`
	RoleID int32 `db:"role_id" json:"role_id"`
}

type GetActiveRolesParams struct {
	Column1 string `db:"column_1" json:"column_1"`
	Limit   int32  `db:"limit" json:"limit"`
	Offset  int32  `db:"offset" json:"offset"`
}

type GetActiveRolesRow struct {
	RoleID     int32            `db:"role_id" json:"role_id"`
	RoleName   string           `db:"role_name" json:"role_name"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type GetRolesParams struct {
	Column1 string `db:"column_1" json:"column_1"`
	Limit   int32  `db:"limit" json:"limit"`
	Offset  int32  `db:"offset" json:"offset"`
}

type GetRolesRow struct {
	RoleID     int32            `db:"role_id" json:"role_id"`
	RoleName   string           `db:"role_name" json:"role_name"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type GetTrashedRolesParams struct {
	Column1 string `db:"column_1" json:"column_1"`
	Limit   int32  `db:"limit" json:"limit"`
	Offset  int32  `db:"offset" json:"offset"`
}

type GetTrashedRolesRow struct {
	RoleID     int32            `db:"role_id" json:"role_id"`
	RoleName   string           `db:"role_name" json:"role_name"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type UpdateRoleParams struct {
	RoleID   int32  `db:"role_id" json:"role_id"`
	RoleName string `db:"role_name" json:"role_name"`
}
