package withdrawgraphqlmapper

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/model"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/withdraw"
)

type WithdrawGraphqlMapper interface {
	ToGraphqlWithdrawAll(res *pb.ApiResponseWithdrawAll) *model.APIResponseWithdrawAll
	ToGraphqlWithdrawDelete(res *pb.ApiResponseWithdrawDelete) *model.APIResponseWithdrawDelete
	ToGraphqlResponseWithdraw(res *pb.ApiResponseWithdraw) *model.APIResponseWithdraw
	ToGraphqlResponseWithdraws(res *pb.ApiResponsesWithdraw) *model.APIResponsesWithdraw
	ToGraphqlResponseWithdrawDeleteAt(res *pb.ApiResponseWithdrawDeleteAt) *model.APIResponseWithdrawDeleteAt
	ToGraphqlResponsePaginationWithdraw(res *pb.ApiResponsePaginationWithdraw) *model.APIResponsePaginationWithdraw
	ToGraphqlResponsePaginationWithdrawDeleteAt(res *pb.ApiResponsePaginationWithdrawDeleteAt) *model.APIResponsePaginationWithdrawDeleteAt
}
