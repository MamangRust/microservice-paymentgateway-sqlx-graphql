package repository

import (
	pbrole "github.com/MamangRust/microservice-payment-gateway-grpc/pb/role"
	pbuserrole "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user_role"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/jmoiron/sqlx"
)

type Repositories interface {
	UserQuery() UserQueryRepository
	UserCommand() UserCommandRepository
	Role() RoleRepository
	UserRole() UserRoleRepository
}

type repositories struct {
	userQuery   UserQueryRepository
	userCommand UserCommandRepository
	role        RoleRepository
	userRole    UserRoleRepository
}

// GuardOptions collects the dependency guards applied to the role/user_role
// adapters built inside NewRepositories. Role reads and user_role writes get
// separate breakers so one failure mode cannot starve the other.
type GuardOptions struct {
	Role     []adapter.GuardOption
	UserRole []adapter.GuardOption
}

// NewRepositories wires the user-owned SQL repositories together with the role
// adapters, which are built here from the raw role/user_role gRPC clients under
// their guards. The user service never touches the roles/user_roles tables
// directly: role reads and user_role writes go through the role service gRPC
// API.
func NewRepositories(
	db *sqlx.DB,
	roleQueryClient pbrole.RoleQueryServiceClient,
	userRoleClient pbuserrole.UserRoleServiceClient,
	guards ...GuardOptions,
) Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	return &repositories{
		userCommand: NewUserCommandRepository(db),
		userQuery:   NewUserQueryRepository(db),
		role:        adapter.NewRoleAdapter(roleQueryClient, g.Role...),
		userRole:    adapter.NewUserRoleAdapter(userRoleClient, g.UserRole...),
	}
}

func (r *repositories) UserQuery() UserQueryRepository {
	return r.userQuery
}
func (r *repositories) UserCommand() UserCommandRepository {
	return r.userCommand
}
func (r *repositories) Role() RoleRepository {
	return r.role
}
func (r *repositories) UserRole() UserRoleRepository {
	return r.userRole
}
