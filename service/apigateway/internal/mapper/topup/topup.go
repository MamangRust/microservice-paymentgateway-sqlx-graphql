package topupgraphqlmapper

import (
	graphqlmapper "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/mapper"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/model"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/topup"
)

type topupGraphqlMapper struct {
}

func NewTopupGraphqlMapper() *topupGraphqlMapper {
	return &topupGraphqlMapper{}
}

func (t *topupGraphqlMapper) ToGraphqlResponseTopup(res *pb.ApiResponseTopup) *model.APIResponseTopup {
	return &model.APIResponseTopup{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponseTopup(res.Data),
	}
}

func (t *topupGraphqlMapper) ToGraphqlResponseTopupDeleteAt(res *pb.ApiResponseTopupDeleteAt) *model.APIResponseTopupDeleteAt {
	return &model.APIResponseTopupDeleteAt{
		Status:  res.Status,
		Message: res.Message,
		Data:    t.mapResponseTopupDeleteAt(res.Data),
	}
}

func (t *topupGraphqlMapper) ToGraphqlTopupAll(res *pb.ApiResponseTopupAll) *model.APIResponseTopupAll {
	return &model.APIResponseTopupAll{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (t *topupGraphqlMapper) ToGraphqlTopupDelete(res *pb.ApiResponseTopupDelete) *model.APIResponseTopupDelete {
	return &model.APIResponseTopupDelete{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (t *topupGraphqlMapper) ToGraphqlResponsePaginationTopup(res *pb.ApiResponsePaginationTopup) *model.APIResponsePaginationTopup {
	return &model.APIResponsePaginationTopup{
		Status:     res.Status,
		Message:    res.Message,
		Data:       t.mapResponsesTopup(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (t *topupGraphqlMapper) ToGraphqlResponsePaginationTopupDeleteAt(res *pb.ApiResponsePaginationTopupDeleteAt) *model.APIResponsePaginationTopupDeleteAt {
	return &model.APIResponsePaginationTopupDeleteAt{
		Status:     res.Status,
		Message:    res.Message,
		Data:       t.mapResponsesTopupDeleteAt(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}


func (t *topupGraphqlMapper) mapResponseTopup(res *pb.TopupResponse) *model.TopupResponse {
	if res == nil { return nil }
	return &model.TopupResponse{ID: res.Id, CardNumber: res.CardNumber, TopupNo: res.TopupNo, TopupAmount: int32(res.TopupAmount), TopupMethod: res.TopupMethod, TopupTime: &res.TopupTime, CreatedAt: res.CreatedAt, UpdatedAt: res.UpdatedAt}
}

func (t *topupGraphqlMapper) mapResponsesTopup(res []*pb.TopupResponse) []*model.TopupResponse {
	result := make([]*model.TopupResponse, 0, len(res))
	for _, r := range res { result = append(result, t.mapResponseTopup(r)) }
	return result
}

func (t *topupGraphqlMapper) mapResponseTopupDeleteAt(res *pb.TopupResponseDeleteAt) *model.TopupResponseDeletedAt {
	if res == nil { return nil }
	return &model.TopupResponseDeletedAt{ID: res.Id, CardNumber: res.CardNumber, TopupNo: res.TopupNo, TopupAmount: int32(res.TopupAmount), TopupMethod: res.TopupMethod, TopupTime: &res.TopupTime, CreatedAt: res.CreatedAt, UpdatedAt: res.UpdatedAt, DeletedAt: graphqlmapper.PointerString(res.DeletedAt)}
}

func (t *topupGraphqlMapper) mapResponsesTopupDeleteAt(res []*pb.TopupResponseDeleteAt) []*model.TopupResponseDeletedAt {
	result := make([]*model.TopupResponseDeletedAt, 0, len(res))
	for _, r := range res { result = append(result, t.mapResponseTopupDeleteAt(r)) }
	return result
}
