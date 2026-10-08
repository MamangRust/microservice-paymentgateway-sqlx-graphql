package cardgraphqlmapper

import (
	graphqlmapper "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/mapper"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/model"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
)

type cardResponseMapper struct {
}

func NewCardResponseMapper() *cardResponseMapper {
	return &cardResponseMapper{}
}

func (s *cardResponseMapper) ToGraphqlResponseCard(res *pb.ApiResponseCard) *model.APIResponseCard {
	return &model.APIResponseCard{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapCardResponse(res.Data),
	}
}

func (s *cardResponseMapper) ToGraphqlResponsePaginationCard(res *pb.ApiResponsePaginationCard) *model.APIResponsePaginationCard {
	return &model.APIResponsePaginationCard{
		Status:     res.Status,
		Message:    res.Message,
		Data:       s.mapCardResponses(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}

func (s *cardResponseMapper) ToGraphqlResponseAll(res *pb.ApiResponseCardAll) *model.APIResponseCardAll {
	return &model.APIResponseCardAll{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (s *cardResponseMapper) ToGraphqlResponseDelete(res *pb.ApiResponseCardDelete) *model.APIResponseCardDelete {
	return &model.APIResponseCardDelete{
		Status:  res.Status,
		Message: res.Message,
	}
}

func (s *cardResponseMapper) ToGraphqlResponseCardDeleteAt(res *pb.ApiResponseCardDeleteAt) *model.APIResponseCardDeleteAt {
	return &model.APIResponseCardDeleteAt{
		Status:  res.Status,
		Message: res.Message,
		Data:    s.mapCardResponseDeleteAt(res.Data),
	}
}

func (s *cardResponseMapper) ToGraphqlResponsePaginationCardDeleteAt(res *pb.ApiResponsePaginationCardDeleteAt) *model.APIResponsePaginationCardDeleteAt {
	return &model.APIResponsePaginationCardDeleteAt{
		Status:     res.Status,
		Message:    res.Message,
		Data:       s.mapCardDeleteAtResponses(res.Data),
		Pagination: graphqlmapper.MapPaginationMeta(res.PaginationMeta),
	}
}


func (s *cardResponseMapper) mapCardResponse(res *pb.CardResponse) *model.CardResponse {
	if res == nil {
		return nil
	}
	return &model.CardResponse{
		ID:           res.Id,
		UserID:       res.UserId,
		CardNumber:   res.CardNumber,
		CardType:     res.CardType,
		ExpireDate:   res.ExpireDate,
		Cvv:          res.Cvv,
		CardProvider: res.CardProvider,
		CreatedAt:    res.CreatedAt,
		UpdatedAt:    res.UpdatedAt,
	}
}

func (s *cardResponseMapper) mapCardResponses(res []*pb.CardResponse) []*model.CardResponse {
	result := make([]*model.CardResponse, 0, len(res))
	for _, r := range res {
		result = append(result, s.mapCardResponse(r))
	}
	return result
}

func (s *cardResponseMapper) mapCardResponseDeleteAt(res *pb.CardResponseDeleteAt) *model.CardResponseDeleteAt {
	if res == nil {
		return nil
	}
	return &model.CardResponseDeleteAt{
		ID:           res.Id,
		UserID:       res.UserId,
		CardNumber:   res.CardNumber,
		CardType:     res.CardType,
		ExpireDate:   res.ExpireDate,
		Cvv:          res.Cvv,
		CardProvider: res.CardProvider,
		CreatedAt:    res.CreatedAt,
		UpdatedAt:    res.UpdatedAt,
		DeletedAt:    graphqlmapper.PointerString(res.DeletedAt),
	}
}

func (s *cardResponseMapper) mapCardDeleteAtResponses(res []*pb.CardResponseDeleteAt) []*model.CardResponseDeleteAt {
	result := make([]*model.CardResponseDeleteAt, 0, len(res))
	for _, r := range res {
		result = append(result, s.mapCardResponseDeleteAt(r))
	}
	return result
}
