package graphqlmapper

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/model"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/common"
)

func MapPaginationMeta(s *pb.PaginationMeta) *model.PaginationMeta {
	return &model.PaginationMeta{
		CurrentPage:  int32(s.CurrentPage),
		PageSize:     int32(s.PageSize),
		TotalRecords: int32(s.TotalRecords),
		TotalPages:   int32(s.TotalPages),
	}
}
