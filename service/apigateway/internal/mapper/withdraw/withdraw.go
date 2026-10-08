package withdrawgraphqlmapper

import (
	graphqlmapper "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/mapper"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/model"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/withdraw"
)

type withdrawGraphqlMapper struct {
}

func NewWithdrawGraphqlMapper() *withdrawGraphqlMapper {
	return &withdrawGraphqlMapper{}
}

func (t *withdrawGraphqlMapper) ToGraphqlWithdrawAll(res *pb.ApiResponseWithdrawAll) *model.APIResponseWithdrawAll {
	return &model.APIResponseWithdrawAll{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (t *withdrawGraphqlMapper) ToGraphqlWithdrawDelete(res *pb.ApiResponseWithdrawDelete) *model.APIResponseWithdrawDelete {
	return &model.APIResponseWithdrawDelete{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (t *withdrawGraphqlMapper) ToGraphqlResponseWithdraw(res *pb.ApiResponseWithdraw) *model.APIResponseWithdraw {
	return &model.APIResponseWithdraw{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponseWithdraw(res.Data),
	}
}

func (t *withdrawGraphqlMapper) ToGraphqlResponseWithdraws(res *pb.ApiResponsesWithdraw) *model.APIResponsesWithdraw {
	return &model.APIResponsesWithdraw{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponsesWithdraw(res.Data),
	}
}

func (t *withdrawGraphqlMapper) ToGraphqlResponseWithdrawDeleteAt(res *pb.ApiResponseWithdrawDeleteAt) *model.APIResponseWithdrawDeleteAt {
	return &model.APIResponseWithdrawDeleteAt{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponseWithdrawDeleteAt(res.Data),
	}
}

func (t *withdrawGraphqlMapper) ToGraphqlResponsePaginationWithdraw(res *pb.ApiResponsePaginationWithdraw) *model.APIResponsePaginationWithdraw {
	return &model.APIResponsePaginationWithdraw{
		Status:     res.Status,
		Message:    res.Message,
		Data:       t.mapResponsesWithdraw(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (t *withdrawGraphqlMapper) ToGraphqlResponsePaginationWithdrawDeleteAt(res *pb.ApiResponsePaginationWithdrawDeleteAt) *model.APIResponsePaginationWithdrawDeleteAt {
	return &model.APIResponsePaginationWithdrawDeleteAt{
		Status:     res.Status,
		Message:    res.Message,
		Data:       t.mapResponsesWithdrawDeleteAt(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}


func (t *withdrawGraphqlMapper) mapResponseWithdraw(res *pb.WithdrawResponse) *model.WithdrawResponse {
	if res == nil { return nil }
	return &model.WithdrawResponse{WithdrawID: res.WithdrawId, WithdrawNo: res.WithdrawNo, CardNumber: res.CardNumber, WithdrawAmount: int32(res.WithdrawAmount), WithdrawTime: res.WithdrawTime, CreatedAt: res.CreatedAt, UpdatedAt: res.UpdatedAt}
}

func (t *withdrawGraphqlMapper) mapResponsesWithdraw(res []*pb.WithdrawResponse) []*model.WithdrawResponse {
	result := make([]*model.WithdrawResponse, 0, len(res))
	for _, r := range res { result = append(result, t.mapResponseWithdraw(r)) }
	return result
}

func (t *withdrawGraphqlMapper) mapResponseWithdrawDeleteAt(res *pb.WithdrawResponseDeleteAt) *model.WithdrawResponseDeletedAt {
	if res == nil { return nil }
	return &model.WithdrawResponseDeletedAt{WithdrawID: res.WithdrawId, WithdrawNo: res.WithdrawNo, CardNumber: res.CardNumber, WithdrawAmount: int32(res.WithdrawAmount), WithdrawTime: res.WithdrawTime, CreatedAt: res.CreatedAt, UpdatedAt: res.UpdatedAt, DeletedAt: graphqlmapper.PointerString(res.DeletedAt)}
}

func (t *withdrawGraphqlMapper) mapResponsesWithdrawDeleteAt(res []*pb.WithdrawResponseDeleteAt) []*model.WithdrawResponseDeletedAt {
	result := make([]*model.WithdrawResponseDeletedAt, 0, len(res))
	for _, r := range res { result = append(result, t.mapResponseWithdrawDeleteAt(r)) }
	return result
}
