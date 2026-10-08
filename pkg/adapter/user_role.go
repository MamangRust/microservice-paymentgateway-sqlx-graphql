package adapter

import (
	"context"

	pbrole "github.com/MamangRust/microservice-payment-gateway-grpc/pb/role"
	pbuserrole "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user_role"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/resilience"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
)

// UserRoleAdapter is the user<->role surface: resolving a user's roles and
// (un)assigning them. It owns the only calls to pb/user_role's UserRoleService.
type UserRoleAdapter interface {
	FindByUserId(ctx context.Context, userID int) ([]*models.Role, error)
	AssignRoleToUser(ctx context.Context, request *requests.CreateUserRoleRequest) (*models.UserRole, error)
	RemoveRoleFromUser(ctx context.Context, request *requests.RemoveUserRoleRequest) error
}

type userRoleGRPCAdapter struct {
	client pbuserrole.UserRoleServiceClient
	guard  *resilience.DependencyGuard
}

func (a *userRoleGRPCAdapter) SetGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

// NewUserRoleAdapter builds the user-role adapter on top of the generated
// UserRoleService client. Passing no options leaves the guard nil.
func NewUserRoleAdapter(client pbuserrole.UserRoleServiceClient, opts ...GuardOption) UserRoleAdapter {
	a := &userRoleGRPCAdapter{client: client}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *userRoleGRPCAdapter) FindByUserId(ctx context.Context, userID int) ([]*models.Role, error) {
	var resp *pbrole.ApiResponsesRole
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.client.FindByUserId(callCtx, &pbuserrole.FindByIdUserRoleRequest{UserId: int32(userID)})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}

	roles := make([]*models.Role, 0, len(resp.Data))
	for _, role := range resp.Data {
		if role == nil {
			continue
		}
		roles = append(roles, mapRoleResponse(role))
	}
	return roles, nil
}

func (a *userRoleGRPCAdapter) AssignRoleToUser(ctx context.Context, request *requests.CreateUserRoleRequest) (*models.UserRole, error) {
	var resp *pbuserrole.ApiResponseUserRole
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.client.AssignRoleToUser(callCtx, &pbuserrole.AssignRoleToUserRequest{
			UserId: int32(request.UserId),
			RoleId: int32(request.RoleId),
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Data == nil {
		return nil, sharedErrors.ErrFailed("assign role to user")
	}
	return &models.UserRole{
		UserRoleID: resp.Data.UserRoleId,
		UserID:     resp.Data.UserId,
		RoleID:     resp.Data.RoleId,
	}, nil
}

func (a *userRoleGRPCAdapter) RemoveRoleFromUser(ctx context.Context, request *requests.RemoveUserRoleRequest) error {
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		_, callErr := a.client.RemoveRoleFromUser(callCtx, &pbuserrole.RemoveRoleFromUserRequest{
			UserId: int32(request.UserId),
			RoleId: int32(request.RoleId),
		})
		return callErr
	})
	if err != nil {
		return sharedErrors.ErrFailed("remove role from user").WithInternal(err)
	}
	return nil
}
