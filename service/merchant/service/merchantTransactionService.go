package service

import (
	"context"

	mencache "github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/redis"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errorhandler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// merchantTransactionDeps holds dependencies for merchant transaction operations.
type merchantTransactionDeps struct {
	TransactionAdapter     adapter.TransactionAdapter
	MerchantQueryRepository repository.MerchantQueryRepository
	Cache                  mencache.MerchantTransactionCache
	Logger                 logger.LoggerInterface
	Observability          observability.TraceLoggerObservability
}

// merchantTransactionService handles merchant transaction operations.
// Transactions are owned by the transaction service; this service reads them
// through the TransactionAdapter gRPC client rather than querying the
// transactions table directly.
type merchantTransactionService struct {
	adapter            adapter.TransactionAdapter
	merchantRepository repository.MerchantQueryRepository
	cache              mencache.MerchantTransactionCache
	logger             logger.LoggerInterface
	observability      observability.TraceLoggerObservability
}

// NewMerchantTransactionService constructs a MerchantTransactionService.
func NewMerchantTransactionService(
	params *merchantTransactionDeps,
) MerchantTransactionService {
	return &merchantTransactionService{
		adapter:            params.TransactionAdapter,
		merchantRepository: params.MerchantQueryRepository,
		cache:              params.Cache,
		logger:             params.Logger,
		observability:      params.Observability,
	}
}

func (s *merchantTransactionService) FindAllTransactions(ctx context.Context, req *requests.FindAllMerchantTransactions) ([]*models.Transaction, *int, error) {
	const method = "FindAllTransactions"
	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))
	defer func() { end(status, "grpc") }()

	if data, total, found := s.cache.GetCacheAllMerchantTransactions(ctx, req); found {
		logSuccess("Successfully retrieved all merchant transactions from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))
		return data, total, nil
	}

	transactions, totalRecords, err := s.adapter.FindAllTransactions(ctx, page, pageSize, search)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*models.Transaction](s.logger, err, method, span, zap.String("search", search))
	}

	// Merchant names are owned by this service, so resolve them locally rather
	// than expecting the transaction service to supply them.
	s.enrichMerchantNames(ctx, transactions)

	s.cache.SetCacheAllMerchantTransactions(ctx, req, transactions, totalRecords)

	logSuccess("Successfully retrieved all merchant transactions", zap.Int("totalRecords", derefInt(totalRecords)), zap.Int("page", page), zap.Int("pageSize", pageSize))

	return transactions, totalRecords, nil
}

func (s *merchantTransactionService) FindAllTransactionsByMerchant(ctx context.Context, req *requests.FindAllMerchantTransactionsById) ([]*models.Transaction, *int, error) {
	const method = "FindAllTransactionsByMerchant"
	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.Int("merchant_id", req.MerchantID), attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))
	defer func() { end(status, "grpc") }()

	if data, total, found := s.cache.GetCacheMerchantTransactions(ctx, req); found {
		logSuccess("Successfully retrieved merchant transactions from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))
		return data, total, nil
	}

	transactions, totalRecords, err := s.adapter.FindAllTransactionsByMerchantId(ctx, req.MerchantID, page, pageSize, search)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*models.Transaction](s.logger, err, method, span, zap.Int("merchant_id", req.MerchantID), zap.String("search", search))
	}

	merchantName := s.resolveMerchantName(ctx, req.MerchantID)
	for _, tx := range transactions {
		if tx != nil {
			tx.MerchantName = merchantName
		}
	}

	s.cache.SetCacheMerchantTransactions(ctx, req, transactions, totalRecords)

	logSuccess("Successfully retrieved merchant transactions", zap.Int("totalRecords", derefInt(totalRecords)), zap.Int("page", page), zap.Int("pageSize", pageSize))

	return transactions, totalRecords, nil
}

func (s *merchantTransactionService) FindAllTransactionsByApikey(ctx context.Context, req *requests.FindAllMerchantTransactionsByApiKey) ([]*models.Transaction, *int, error) {
	const method = "FindAllTransactionsByApikey"
	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.String("api_key", req.ApiKey), attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))
	defer func() { end(status, "grpc") }()

	if data, total, found := s.cache.GetCacheMerchantTransactionApikey(ctx, req); found {
		logSuccess("Successfully retrieved merchant transactions by apikey from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))
		return data, total, nil
	}

	merchant, err := s.merchantRepository.FindByApiKey(ctx, req.ApiKey)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*models.Transaction](s.logger, err, method, span, zap.String("api_key", req.ApiKey))
	}

	transactions, totalRecords, err := s.adapter.FindAllTransactionsByMerchantId(ctx, int(merchant.MerchantID), page, pageSize, search)
	if err != nil {
		status = "error"
		return errorhandler.HandlerErrorPagination[[]*models.Transaction](s.logger, err, method, span, zap.String("api_key", req.ApiKey))
	}

	// The merchant was already resolved by API key; reuse its name.
	for _, tx := range transactions {
		if tx != nil {
			tx.MerchantName = merchant.Name
		}
	}

	s.cache.SetCacheMerchantTransactionApikey(ctx, req, transactions, totalRecords)

	logSuccess("Successfully retrieved merchant transactions by apikey", zap.Int("totalRecords", derefInt(totalRecords)), zap.Int("page", page), zap.Int("pageSize", pageSize))

	return transactions, totalRecords, nil
}

func (s *merchantTransactionService) normalizePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize
}

func derefInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

// resolveMerchantName looks up a merchant's name from this service's own
// merchants table (the merchant domain is owned here).
func (s *merchantTransactionService) resolveMerchantName(ctx context.Context, merchantID int) string {
	merchant, err := s.merchantRepository.FindByMerchantId(ctx, merchantID)
	if err != nil || merchant == nil {
		return ""
	}
	return merchant.Name
}

// enrichMerchantNames fills MerchantName on each transaction using the
// merchant records owned by this service. Names are resolved once per distinct
// merchant ID to keep the lookup count bounded by the page.
func (s *merchantTransactionService) enrichMerchantNames(ctx context.Context, transactions []*models.Transaction) {
	names := make(map[int]string)
	for _, tx := range transactions {
		if tx == nil {
			continue
		}
		name, ok := names[tx.MerchantID]
		if !ok {
			name = s.resolveMerchantName(ctx, tx.MerchantID)
			names[tx.MerchantID] = name
		}
		tx.MerchantName = name
	}
}
