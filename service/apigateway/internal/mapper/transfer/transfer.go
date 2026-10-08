package transfergraphqlmapper

import (
	graphqlmapper "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/mapper"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/model"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transfer"
)

type transferGraphqlMapper struct {
}

func NewTransferGraphqlMapper() *transferGraphqlMapper {
	return &transferGraphqlMapper{}
}

func (t *transferGraphqlMapper) ToGraphqlTransferAll(res *pb.ApiResponseTransferAll) *model.APIResponseTransferAll {
	return &model.APIResponseTransferAll{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (t *transferGraphqlMapper) ToGraphqlTransferDelete(res *pb.ApiResponseTransferDelete) *model.APIResponseTransferDelete {
	return &model.APIResponseTransferDelete{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (t *transferGraphqlMapper) ToGraphqlResponseTransfer(res *pb.ApiResponseTransfer) *model.APIResponseTransfer {
	return &model.APIResponseTransfer{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponseTransfer(res.Data),
	}
}

func (t *transferGraphqlMapper) ToGraphqlResponseTransfers(res *pb.ApiResponseTransfers) *model.APIResponseTransfers {
	return &model.APIResponseTransfers{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesTransfer(res.Data),
	}
}

func (t *transferGraphqlMapper) ToGraphqlResponseTransferDeleteAt(res *pb.ApiResponseTransferDeleteAt) *model.APIResponseTransferDeleteAt {
	return &model.APIResponseTransferDeleteAt{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponseTransferDeleteAt(res.Data),
	}
}

func (t *transferGraphqlMapper) ToGraphqlResponsePaginationTransfer(res *pb.ApiResponsePaginationTransfer) *model.APIResponsePaginationTransfer {
	return &model.APIResponsePaginationTransfer{
		Status:     res.Status,
		Message:    res.Message,
		Data:       t.mapResponsesTransfer(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (t *transferGraphqlMapper) ToGraphqlResponsePaginationTransferDeleteAt(res *pb.ApiResponsePaginationTransferDeleteAt) *model.APIResponsePaginationTransferDeleteAt {
	return &model.APIResponsePaginationTransferDeleteAt{
		Status:     res.Status,
		Message:    res.Message,
		Data:       t.mapResponsesTransferDeleteAt(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}


func (t *transferGraphqlMapper) mapResponseTransfer(res *pb.TransferResponse) *model.TransferResponse {
	if res == nil { return nil }
	return &model.TransferResponse{ID: res.Id, TransferNo: res.TransferNo, TransferFrom: res.TransferFrom, TransferTo: res.TransferTo, TransferAmount: int32(res.TransferAmount), TransferTime: res.TransferTime, CreatedAt: res.CreatedAt, UpdatedAt: res.UpdatedAt}
}

func (t *transferGraphqlMapper) mapResponsesTransfer(res []*pb.TransferResponse) []*model.TransferResponse {
	result := make([]*model.TransferResponse, 0, len(res))
	for _, r := range res { result = append(result, t.mapResponseTransfer(r)) }
	return result
}

func (t *transferGraphqlMapper) mapResponseTransferDeleteAt(res *pb.TransferResponseDeleteAt) *model.TransferResponseDeletedAt {
	if res == nil { return nil }
	return &model.TransferResponseDeletedAt{ID: res.Id, TransferNo: res.TransferNo, TransferFrom: res.TransferFrom, TransferTo: res.TransferTo, TransferAmount: int32(res.TransferAmount), TransferTime: res.TransferTime, CreatedAt: res.CreatedAt, UpdatedAt: res.UpdatedAt, DeletedAt: graphqlmapper.PointerString(res.DeletedAt)}
}

func (t *transferGraphqlMapper) mapResponsesTransferDeleteAt(res []*pb.TransferResponseDeleteAt) []*model.TransferResponseDeletedAt {
	result := make([]*model.TransferResponseDeletedAt, 0, len(res))
	for _, r := range res { result = append(result, t.mapResponseTransferDeleteAt(r)) }
	return result
}
