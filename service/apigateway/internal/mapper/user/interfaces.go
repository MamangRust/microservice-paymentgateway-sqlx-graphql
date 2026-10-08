package usergraphqlmapper

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/model"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
)

type UserGraphqlMapper interface {
	ToGraphqlResponseUser(res *pb.ApiResponseUser) *model.APIResponseUser
	ToGraphqlResponseUserDeleteAt(res *pb.ApiResponseUserDeleteAt) *model.APIResponseUserDeleteAt
	ToGraphqlResponseUsers(res *pb.ApiResponsesUser) *model.APIResponsesUser
	ToGraphqlResponseUserDelete(res *pb.ApiResponseUserDelete) *model.APIResponseUserDelete
	ToGraphqlResponseUserAll(res *pb.ApiResponseUserAll) *model.APIResponseUserAll
	ToGraphqlResponsePaginationUser(res *pb.ApiResponsePaginationUser) *model.APIResponsePaginationUser
	ToGraphqlResponsePaginationUserDeleteAt(res *pb.ApiResponsePaginationUserDeleteAt) *model.APIResponsePaginationUserDeleteAt
}
