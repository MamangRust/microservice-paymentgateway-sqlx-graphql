package cardgraphqlmapper

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/model"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
)

type CardGraphqlMapper interface {
	ToGraphqlResponsePaginationCard(res *pb.ApiResponsePaginationCard) *model.APIResponsePaginationCard
	ToGraphqlResponseCard(res *pb.ApiResponseCard) *model.APIResponseCard
	ToGraphqlResponseAll(res *pb.ApiResponseCardAll) *model.APIResponseCardAll
	ToGraphqlResponseDelete(res *pb.ApiResponseCardDelete) *model.APIResponseCardDelete
	ToGraphqlResponseCardDeleteAt(res *pb.ApiResponseCardDeleteAt) *model.APIResponseCardDeleteAt
	ToGraphqlResponsePaginationCardDeleteAt(res *pb.ApiResponsePaginationCardDeleteAt) *model.APIResponsePaginationCardDeleteAt
}
