package merchantgraphqlmapper

import (
	graphqlmapper "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/mapper"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/model"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant"
)

type merchantResponse struct{}

func NewMerchantResponseMapper() *merchantResponse {
	return &merchantResponse{}
}

func (m *merchantResponse) ToGraphqlResponseMerchant(res *pb.ApiResponseMerchant) *model.APIResponseMerchant {
	return &model.APIResponseMerchant{
		Status:  res.Status,
		Message: res.Message,
		Data:    m.mapMerchantResponse(res.Data),
	}
}

func (m *merchantResponse) ToGraphqlResponsePaginationMerchant(res *pb.ApiResponsePaginationMerchant) *model.APIResponsePaginationMerchant {
	return &model.APIResponsePaginationMerchant{
		Status:     res.Status,
		Message:    res.Message,
		Data:       m.mapMerchantResponses(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (m *merchantResponse) ToGraphqlResponsesMerchant(res *pb.ApiResponsesMerchant) *model.APIResponsesMerchant {
	return &model.APIResponsesMerchant{
		Status:  res.Status,
		Message: res.Message,
		Data:    m.mapMerchantResponses(res.Data),
	}
}

func (m *merchantResponse) ToGraphqlResponseMerchantDeleteAt(res *pb.ApiResponseMerchantDeleteAt) *model.APIResponseMerchantDeleteAt {
	return &model.APIResponseMerchantDeleteAt{
		Status:  res.Status,
		Message: res.Message,
		Data:    m.mapMerchantResponseDeleteAt(res.Data),
	}
}

func (m *merchantResponse) ToGraphqlResponsePaginationMerchantDeleteAt(res *pb.ApiResponsePaginationMerchantDeleteAt) *model.APIResponsePaginationMerchantDeleteAt {
	return &model.APIResponsePaginationMerchantDeleteAt{
		Status:     res.Status,
		Message:    res.Message,
		Data:       m.mapMerchantResponsesDeleteAt(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}



func (m *merchantResponse) mapMerchantResponse(res *pb.MerchantResponse) *model.MerchantResponse {
	if res == nil { return nil }
	return &model.MerchantResponse{ID: res.Id, Name: res.Name, APIKey: res.ApiKey, Status: res.Status, UserID: res.UserId, CreatedAt: res.CreatedAt, UpdatedAt: res.UpdatedAt}
}

func (m *merchantResponse) mapMerchantResponses(res []*pb.MerchantResponse) []*model.MerchantResponse {
	result := make([]*model.MerchantResponse, 0, len(res))
	for _, r := range res { result = append(result, m.mapMerchantResponse(r)) }
	return result
}

func (m *merchantResponse) mapMerchantResponseDeleteAt(res *pb.MerchantResponseDeleteAt) *model.MerchantResponseDeletedAt {
	if res == nil { return nil }
	return &model.MerchantResponseDeletedAt{ID: res.Id, Name: res.Name, APIKey: res.ApiKey, Status: res.Status, UserID: res.UserId, CreatedAt: res.CreatedAt, UpdatedAt: res.UpdatedAt, DeletedAt: graphqlmapper.PointerString(res.DeletedAt)}
}

func (m *merchantResponse) mapMerchantResponsesDeleteAt(res []*pb.MerchantResponseDeleteAt) []*model.MerchantResponseDeletedAt {
	result := make([]*model.MerchantResponseDeletedAt, 0, len(res))
	for _, r := range res { result = append(result, m.mapMerchantResponseDeleteAt(r)) }
	return result
}

func (m *merchantResponse) ToGraphqlMerchantDeleteAll(res *pb.ApiResponseMerchantDelete) *model.APIResponseMerchantDelete {
	return &model.APIResponseMerchantDelete{Status: res.Status, Message: res.Message}
}

func (m *merchantResponse) ToGraphqlMerchantAll(res *pb.ApiResponseMerchantAll) *model.APIResponseMerchantAll {
	return &model.APIResponseMerchantAll{Status: res.Status, Message: res.Message}
}
