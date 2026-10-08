package adapter

import (
	"context"

	pbrole "github.com/MamangRust/microservice-payment-gateway-grpc/pb/role"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/resilience"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
)

// RoleAdapter is the read-only role surface. Role assignment now lives in the
// user_role service (UserRoleAdapter), so this adapter only exposes lookups.
type RoleAdapter interface {
	FindById(ctx context.Context, roleID int) (*models.Role, error)
	FindByName(ctx context.Context, name string) (*models.Role, error)
}

type roleGRPCAdapter struct {
	queryClient pbrole.RoleQueryServiceClient
	guard       *resilience.DependencyGuard
}

func (a *roleGRPCAdapter) SetGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

// NewRoleAdapter builds the role adapter on top of the generated RoleQueryService
// client. Passing no options leaves the guard nil (DependencyGuard.Call becomes a
// passthrough).
func NewRoleAdapter(queryClient pbrole.RoleQueryServiceClient, opts ...GuardOption) RoleAdapter {
	a := &roleGRPCAdapter{queryClient: queryClient}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *roleGRPCAdapter) FindById(ctx context.Context, roleID int) (*models.Role, error) {
	var resp *pbrole.ApiResponseRole
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.queryClient.FindByIdRole(callCtx, &pbrole.FindByIdRoleRequest{RoleId: int32(roleID)})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Data == nil {
		return nil, sharedErrors.ErrRoleNotFound
	}
	return mapRoleResponse(resp.Data), nil
}

func (a *roleGRPCAdapter) FindByName(ctx context.Context, name string) (*models.Role, error) {
	var resp *pbrole.ApiResponseRole
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.queryClient.FindByNameRole(callCtx, &pbrole.FindByNameRoleRequest{Name: name})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Data == nil {
		return nil, sharedErrors.ErrRoleNotFound
	}
	return mapRoleResponse(resp.Data), nil
}

func mapRoleResponse(r *pbrole.RoleResponse) *models.Role {
	if r == nil {
		return nil
	}
	return &models.Role{
		RoleID:    r.Id,
		RoleName:  r.Name,
		CreatedAt: parseTime(r.CreatedAt),
		UpdatedAt: parseTime(r.UpdatedAt),
	}
}
