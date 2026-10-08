package transfergraphqlmapper

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/model"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transfer"
)

type TransferGraphqlMapper interface {
	ToGraphqlTransferAll(res *pb.ApiResponseTransferAll) *model.APIResponseTransferAll
	ToGraphqlTransferDelete(res *pb.ApiResponseTransferDelete) *model.APIResponseTransferDelete
	ToGraphqlResponseTransfer(res *pb.ApiResponseTransfer) *model.APIResponseTransfer
	ToGraphqlResponseTransfers(res *pb.ApiResponseTransfers) *model.APIResponseTransfers
	ToGraphqlResponseTransferDeleteAt(res *pb.ApiResponseTransferDeleteAt) *model.APIResponseTransferDeleteAt
	ToGraphqlResponsePaginationTransfer(res *pb.ApiResponsePaginationTransfer) *model.APIResponsePaginationTransfer
	ToGraphqlResponsePaginationTransferDeleteAt(res *pb.ApiResponsePaginationTransferDeleteAt) *model.APIResponsePaginationTransferDeleteAt
}
