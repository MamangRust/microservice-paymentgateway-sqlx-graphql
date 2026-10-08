package merchantgraphqlmapper

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/model"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant"
)

type MerchantGraphqlMapper interface {
	ToGraphqlResponseMerchant(res *pb.ApiResponseMerchant) *model.APIResponseMerchant
	ToGraphqlResponsesMerchant(res *pb.ApiResponsesMerchant) *model.APIResponsesMerchant
	ToGraphqlResponsePaginationMerchant(res *pb.ApiResponsePaginationMerchant) *model.APIResponsePaginationMerchant
	ToGraphqlResponseMerchantDeleteAt(res *pb.ApiResponseMerchantDeleteAt) *model.APIResponseMerchantDeleteAt
	ToGraphqlResponsePaginationMerchantDeleteAt(res *pb.ApiResponsePaginationMerchantDeleteAt) *model.APIResponsePaginationMerchantDeleteAt
	ToGraphqlMerchantDeleteAll(res *pb.ApiResponseMerchantDelete) *model.APIResponseMerchantDelete
	ToGraphqlMerchantAll(res *pb.ApiResponseMerchantAll) *model.APIResponseMerchantAll
}
