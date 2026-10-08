package handler

import (
	"context"

	pbrole "github.com/MamangRust/microservice-payment-gateway-grpc/pb/role"
	pbuserrole "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user_role"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RoleQueryHandlerGrpc interface {
	pbrole.RoleQueryServiceServer

	FindAllRole(ctx context.Context, req *pbrole.FindAllRoleRequest) (*pbrole.ApiResponsePaginationRole, error)
	FindByIdRole(ctx context.Context, req *pbrole.FindByIdRoleRequest) (*pbrole.ApiResponseRole, error)
	FindByActive(ctx context.Context, req *pbrole.FindAllRoleRequest) (*pbrole.ApiResponsePaginationRoleDeleteAt, error)
	FindByTrashed(ctx context.Context, req *pbrole.FindAllRoleRequest) (*pbrole.ApiResponsePaginationRoleDeleteAt, error)
	FindByNameRole(ctx context.Context, req *pbrole.FindByNameRoleRequest) (*pbrole.ApiResponseRole, error)
}

type RoleCommandHandlerGrpc interface {
	pbrole.RoleCommandServiceServer

	CreateRole(ctx context.Context, req *pbrole.CreateRoleRequest) (*pbrole.ApiResponseRole, error)
	UpdateRole(ctx context.Context, req *pbrole.UpdateRoleRequest) (*pbrole.ApiResponseRole, error)
	TrashedRole(ctx context.Context, req *pbrole.FindByIdRoleRequest) (*pbrole.ApiResponseRoleDeleteAt, error)
	RestoreRole(ctx context.Context, req *pbrole.FindByIdRoleRequest) (*pbrole.ApiResponseRoleDeleteAt, error)
	DeleteRolePermanent(ctx context.Context, req *pbrole.FindByIdRoleRequest) (*pbrole.ApiResponseRoleDelete, error)
	RestoreAllRole(ctx context.Context, req *emptypb.Empty) (*pbrole.ApiResponseRoleAll, error)
	DeleteAllRolePermanent(ctx context.Context, req *emptypb.Empty) (*pbrole.ApiResponseRoleAll, error)
}

// UserRoleHandlerGrpc owns the user<->role surface: resolving a user's roles and
// (un)assigning them. It is served by the role service alongside the role
// query/command handlers.
type UserRoleHandlerGrpc interface {
	pbuserrole.UserRoleServiceServer

	FindByUserId(ctx context.Context, req *pbuserrole.FindByIdUserRoleRequest) (*pbrole.ApiResponsesRole, error)
	AssignRoleToUser(ctx context.Context, req *pbuserrole.AssignRoleToUserRequest) (*pbuserrole.ApiResponseUserRole, error)
	RemoveRoleFromUser(ctx context.Context, req *pbuserrole.RemoveRoleFromUserRequest) (*emptypb.Empty, error)
}
