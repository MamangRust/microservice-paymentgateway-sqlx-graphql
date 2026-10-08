package transactiongraphqlmapper

import (
	graphqlmapper "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/mapper"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/model"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transaction"
)

type transactionGraphqlMapper struct{}

func NewTransactionGraphqlMapper() *transactionGraphqlMapper {
	return &transactionGraphqlMapper{}
}

func (t *transactionGraphqlMapper) ToGraphqlTransactionAll(res *pb.ApiResponseTransactionAll) *model.APIResponseTransactionAll {
	return &model.APIResponseTransactionAll{Status: res.Status, Message: res.Message}
}

func (t *transactionGraphqlMapper) ToGraphqlTransactionDelete(res *pb.ApiResponseTransactionDelete) *model.APIResponseTransactionDelete {
	return &model.APIResponseTransactionDelete{Status: res.Status, Message: res.Message}
}

func (t *transactionGraphqlMapper) ToGraphqlPaginationTransaction(res *pb.ApiResponsePaginationTransaction) *model.APIResponsePaginationTransaction {
	return &model.APIResponsePaginationTransaction{Status: res.Status, Message: res.Message, Data: t.mapTransactions(res.Data), Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta)}
}

func (t *transactionGraphqlMapper) ToGraphqlPaginationTransactionDeleteAt(res *pb.ApiResponsePaginationTransactionDeleteAt) *model.APIResponsePaginationTransactionDeleteAt {
	return &model.APIResponsePaginationTransactionDeleteAt{Status: res.Status, Message: res.Message, Data: t.mapTransactionDeleteAts(res.Data), Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta)}
}

func (t *transactionGraphqlMapper) ToGraphqlResponseTransaction(res *pb.ApiResponseTransaction) *model.APIResponseTransaction {
	return &model.APIResponseTransaction{Status: res.Status, Message: res.Message, Data: t.mapTransaction(res.Data)}
}

func (t *transactionGraphqlMapper) ToGraphqlResponseTransactions(res *pb.ApiResponseTransactions) *model.APIResponseTransactions {
	return &model.APIResponseTransactions{Status: res.Status, Message: res.Message, Data: t.mapTransactions(res.Data)}
}

func (t *transactionGraphqlMapper) ToGraphqlResponseTransactionDeleteAt(res *pb.ApiResponseTransactionDeleteAt) *model.APIResponseTransactionDeleteAt {
	return &model.APIResponseTransactionDeleteAt{Status: res.Status, Message: res.Message, Data: t.mapTransactionDeleteAt(res.Data)}
}

func (t *transactionGraphqlMapper) mapTransaction(res *pb.TransactionResponse) *model.TransactionResponse {
	if res == nil {
		return nil
	}
	return &model.TransactionResponse{ID: res.Id, CardNumber: res.CardNumber, TransactionNo: res.TransactionNo, Amount: int32(res.Amount), PaymentMethod: res.PaymentMethod, MerchantID: res.MerchantId, TransactionTime: res.TransactionTime, CreatedAt: res.CreatedAt, UpdatedAt: res.UpdatedAt}
}

func (t *transactionGraphqlMapper) mapTransactions(res []*pb.TransactionResponse) []*model.TransactionResponse {
	result := make([]*model.TransactionResponse, 0, len(res))
	for _, r := range res {
		result = append(result, t.mapTransaction(r))
	}
	return result
}

func (t *transactionGraphqlMapper) mapTransactionDeleteAt(res *pb.TransactionResponseDeleteAt) *model.TransactionResponseDeletedAt {
	if res == nil {
		return nil
	}
	return &model.TransactionResponseDeletedAt{ID: res.Id, CardNumber: res.CardNumber, TransactionNo: res.TransactionNo, Amount: int32(res.Amount), PaymentMethod: res.PaymentMethod, MerchantID: res.MerchantId, TransactionTime: res.TransactionTime, CreatedAt: res.CreatedAt, UpdatedAt: res.UpdatedAt, DeletedAt: graphqlmapper.PointerString(res.DeletedAt)}
}

func (t *transactionGraphqlMapper) mapTransactionDeleteAts(res []*pb.TransactionResponseDeleteAt) []*model.TransactionResponseDeletedAt {
	result := make([]*model.TransactionResponseDeletedAt, 0, len(res))
	for _, r := range res {
		result = append(result, t.mapTransactionDeleteAt(r))
	}
	return result
}
