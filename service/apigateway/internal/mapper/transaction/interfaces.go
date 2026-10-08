package transactiongraphqlmapper

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/model"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transaction"
)

type TransactionGraphqlMapper interface {
	ToGraphqlTransactionAll(res *pb.ApiResponseTransactionAll) *model.APIResponseTransactionAll
	ToGraphqlTransactionDelete(res *pb.ApiResponseTransactionDelete) *model.APIResponseTransactionDelete
	ToGraphqlPaginationTransaction(res *pb.ApiResponsePaginationTransaction) *model.APIResponsePaginationTransaction
	ToGraphqlPaginationTransactionDeleteAt(res *pb.ApiResponsePaginationTransactionDeleteAt) *model.APIResponsePaginationTransactionDeleteAt
	ToGraphqlResponseTransaction(res *pb.ApiResponseTransaction) *model.APIResponseTransaction
	ToGraphqlResponseTransactions(res *pb.ApiResponseTransactions) *model.APIResponseTransactions
	ToGraphqlResponseTransactionDeleteAt(res *pb.ApiResponseTransactionDeleteAt) *model.APIResponseTransactionDeleteAt
}
