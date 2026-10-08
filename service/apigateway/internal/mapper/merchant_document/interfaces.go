package merchantdocumentgraphqlmapper

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/model"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant_document"
)

type MerchantDocumentGraphqlMapper interface {
	ToGraphqlResponseMerchantDocument(res *pb.ApiResponseMerchantDocument) *model.APIResponseMerchantDocument
	ToGraphqlResponseMerchantDocumentDeleteAt(res *pb.ApiResponseMerchantDocumentDeleteAt) *model.APIResponseMerchantDocumentDeleteAt

	ToGraphqlResponseDelete(res *pb.ApiResponseMerchantDocumentDelete) *model.APIResponseMerchantDocumentDelete
	ToGraphqlResponseAll(res *pb.ApiResponseMerchantDocumentAll) *model.APIResponseMerchantDocumentAll
	ToGraphqlResponsePaginationMerchantDocument(res *pb.ApiResponsePaginationMerchantDocument) *model.APIResponsePaginationMerchantDocument
	ToGraphqlResponsePaginationMerchantDocumentDeleteAt(res *pb.ApiResponsePaginationMerchantDocumentAt) *model.APIResponsePaginationMerchantDocumentAt
}
