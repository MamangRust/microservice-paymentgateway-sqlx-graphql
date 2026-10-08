package handler

import (
	"context"

	pbrole "github.com/MamangRust/microservice-payment-gateway-grpc/pb/role"
	pbuserrole "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user_role"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/role/service"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	role_errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors/role_errors/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// userRoleHandleGrpc serves the UserRoleService: it owns user<->role lookups and
// (un)assignment, which used to live on the role query/command services.
type userRoleHandleGrpc struct {
	pbuserrole.UnimplementedUserRoleServiceServer
	roleQuery   service.RoleQueryService
	roleCommand service.RoleCommandService
}

func NewUserRoleHandleGrpc(roleQuery service.RoleQueryService, roleCommand service.RoleCommandService) UserRoleHandlerGrpc {
	return &userRoleHandleGrpc{
		roleQuery:   roleQuery,
		roleCommand: roleCommand,
	}
}

func (s *userRoleHandleGrpc) FindByUserId(ctx context.Context, req *pbuserrole.FindByIdUserRoleRequest) (*pbrole.ApiResponsesRole, error) {
	userID := int(req.GetUserId())

	if userID == 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}

	roles, err := s.roleQuery.FindByUserId(ctx, userID)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoRoles := make([]*pbrole.RoleResponse, 0, len(roles))
	for _, role := range roles {
		if role == nil {
			continue
		}
		protoRoles = append(protoRoles, &pbrole.RoleResponse{
			Id:        int32(role.RoleID),
			Name:      role.RoleName,
			CreatedAt: role.CreatedAt.Time.Format("2006-01-02"),
			UpdatedAt: role.UpdatedAt.Time.Format("2006-01-02"),
		})
	}

	return &pbrole.ApiResponsesRole{
		Status:  "success",
		Message: "Successfully fetched role by user id",
		Data:    protoRoles,
	}, nil
}

func (s *userRoleHandleGrpc) AssignRoleToUser(ctx context.Context, req *pbuserrole.AssignRoleToUserRequest) (*pbuserrole.ApiResponseUserRole, error) {
	userID := int(req.GetUserId())
	roleID := int(req.GetRoleId())

	if userID == 0 || roleID == 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}

	if _, err := s.roleCommand.CreateUserRole(ctx, userID, roleID); err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pbuserrole.ApiResponseUserRole{
		Status:  "success",
		Message: "Successfully associated role with user",
		Data: &pbuserrole.UserRoleResponse{
			UserId: int32(userID),
			RoleId: int32(roleID),
		},
	}, nil
}

func (s *userRoleHandleGrpc) RemoveRoleFromUser(ctx context.Context, req *pbuserrole.RemoveRoleFromUserRequest) (*emptypb.Empty, error) {
	userID := int(req.GetUserId())
	roleID := int(req.GetRoleId())

	if userID == 0 || roleID == 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}

	if _, err := s.roleCommand.DeleteUserRole(ctx, userID, roleID); err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &emptypb.Empty{}, nil
}
