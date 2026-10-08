package repository

import (
	"context"

	adapter "github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/auth/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

// UserRepository defines the data access layer for user-related operations.
//
// UserRepository is the auth-facing user surface, satisfied by the shared
// user adapter (pkg/adapter/user_grpc.go).
//
//go:generate mockgen -source=interfaces.go -destination=mocks/mock.go
type UserRepository = adapter.AuthUserAdapter

type ResetTokenRepository interface {
	FindByToken(ctx context.Context, code string) (*db.GetResetTokenRow, error)

	CreateResetToken(ctx context.Context, req *requests.CreateResetTokenRequest) (*db.CreateResetTokenRow, error)

	DeleteResetToken(ctx context.Context, user_id int32) error
}

type RefreshTokenRepository interface {
	FindByToken(ctx context.Context, token string) (*db.RefreshToken, error)

	FindByUserId(ctx context.Context, user_id int) (*db.RefreshToken, error)

	CreateRefreshToken(ctx context.Context, req *requests.CreateRefreshToken) (*db.RefreshToken, error)

	UpdateRefreshToken(ctx context.Context, req *requests.UpdateRefreshToken) (*db.RefreshToken, error)

	DeleteRefreshToken(ctx context.Context, token string) error

	DeleteRefreshTokenByUserId(ctx context.Context, user_id int) error
}

type UserRoleRepository = adapter.UserRoleAdapter

type RoleRepository = adapter.RoleAdapter
