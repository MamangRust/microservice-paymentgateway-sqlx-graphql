package topupgraphqlmapper

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/model"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/topup"
)

type TopupGraphqlMapper interface {
	ToGraphqlResponseTopup(res *pb.ApiResponseTopup) *model.APIResponseTopup
	ToGraphqlResponseTopupDeleteAt(res *pb.ApiResponseTopupDeleteAt) *model.APIResponseTopupDeleteAt
	ToGraphqlTopupAll(res *pb.ApiResponseTopupAll) *model.APIResponseTopupAll
	ToGraphqlTopupDelete(res *pb.ApiResponseTopupDelete) *model.APIResponseTopupDelete
	ToGraphqlResponsePaginationTopup(res *pb.ApiResponsePaginationTopup) *model.APIResponsePaginationTopup
	ToGraphqlResponsePaginationTopupDeleteAt(res *pb.ApiResponsePaginationTopupDeleteAt) *model.APIResponsePaginationTopupDeleteAt
}
