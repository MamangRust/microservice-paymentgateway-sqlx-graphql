package repository

import (
	pbrole "github.com/MamangRust/microservice-payment-gateway-grpc/pb/role"
	pbuser "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
	pbuserrole "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user_role"
	adapter "github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	User         UserRepository
	RefreshToken RefreshTokenRepository
	UserRole     UserRoleRepository
	Role         RoleRepository
	ResetToken   ResetTokenRepository
}

// GuardOptions collects the dependency guards applied to the user/role/user_role
// adapters built inside NewRepositories. Each dependency gets its own breaker so
// one failing dependency cannot starve the others.
type GuardOptions struct {
	User     []adapter.GuardOption
	Role     []adapter.GuardOption
	UserRole []adapter.GuardOption
}

// RepositoriesDeps carries the raw gRPC clients and guards NewRepositories uses
// to build the user/role/user_role adapters. The auth service never touches the
// users/roles/user_roles tables directly: every cross-domain read/write goes
// through those gRPC APIs under the supplied guards.
type RepositoriesDeps struct {
	DB                *sqlx.DB
	UserQueryClient   pbuser.UserQueryServiceClient
	UserCommandClient pbuser.UserCommandServiceClient
	RoleQueryClient   pbrole.RoleQueryServiceClient
	UserRoleClient    pbuserrole.UserRoleServiceClient
	Guards            GuardOptions
}

func NewRepositories(deps *RepositoriesDeps) *Repositories {
	g := deps.Guards

	return &Repositories{
		User:         adapter.NewAuthUserAdapter(deps.UserQueryClient, deps.UserCommandClient, g.User...),
		UserRole:     adapter.NewUserRoleAdapter(deps.UserRoleClient, g.UserRole...),
		Role:         adapter.NewRoleAdapter(deps.RoleQueryClient, g.Role...),
		RefreshToken: NewRefreshTokenRepository(deps.DB),
		ResetToken:   NewResetTokenRepository(deps.DB),
	}
}
