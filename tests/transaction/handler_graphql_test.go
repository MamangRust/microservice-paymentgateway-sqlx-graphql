package transaction_test

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transaction"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/testhelper"
	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	merchant_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/repository"
	saldo_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transaction/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transaction/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transaction/service"
	user_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"
	tests "github.com/MamangRust/microservice-payment-gateway-test"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TransactionGraphqlTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	dbPool      *sqlx.DB
	redisClient *redis.Client
	grpcServer  *grpc.Server
	conn        *grpc.ClientConn
	gql         *tests.GraphQLClient
	txClient    pb.TransactionCommandServiceClient

	customerCard string
	merchantID   int
	merchantKey  string
	transactionID int
}

func (s *TransactionGraphqlTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card", "merchant", "saldo", "transaction"))

	pool, err := sqlx.ConnectContext(s.ts.Ctx, "pgx", s.ts.DBURL)
	s.Require().NoError(err)
	s.dbPool = pool

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	userRepos := user_repo.NewRepositories(pool, nil, nil)
	cardRepos := card_repo.NewRepositories(pool, nil)
	saldoRepos := saldo_repo.NewRepositories(pool, nil, nil)
	merchantRepos := merchant_repo.NewRepositories(pool, nil, nil)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	transactionService := service.NewService(&service.Deps{
		Kafka:            nil,
		Repositories:     repository.NewRepositories(&repository.Deps{DB: pool}),
		MerchantAdapter:  s.ts.MerchantAdapter,
		CardAdapter:      s.ts.CardAdapter,
		SaldoAdapter:     s.ts.SaldoAdapter,
		Logger:           log,
		Cache:            cacheStore,
		AISecurityClient: nil,
	})

	// Customer: owns the card that is debited.
	customer, err := userRepos.UserCommand().CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Graphql",
		LastName:  "Customer",
		Email:     "graphql.customer@example.com",
		Password:  "password123",
	})
	s.Require().NoError(err)
	cCard, err := cardRepos.CardCommand.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID:       int(customer.UserID),
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(1, 0, 0),
		CVV:          "123",
		CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.customerCard = cCard.CardNumber
	_, err = saldoRepos.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{
		CardNumber:   s.customerCard,
		TotalBalance: 1000000,
	})
	s.Require().NoError(err)

	// Merchant: owner user + active merchant + the merchant's own card.
	owner, err := userRepos.UserCommand().CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Graphql",
		LastName:  "MerchantOwner",
		Email:     "graphql.merchantowner@example.com",
		Password:  "password123",
	})
	s.Require().NoError(err)
	merchant, err := merchantRepos.CreateMerchant(context.Background(), &requests.CreateMerchantRequest{
		UserID: int(owner.UserID),
		Name:   "Graphql Transaction Merchant",
	})
	s.Require().NoError(err)
	s.merchantID = int(merchant.MerchantID)
	_, err = merchantRepos.UpdateMerchantStatus(context.Background(), &requests.UpdateMerchantStatusRequest{
		MerchantID: &s.merchantID,
		Status:     "active",
	})
	s.Require().NoError(err)
	mFull, err := merchantRepos.FindByMerchantId(context.Background(), s.merchantID)
	s.Require().NoError(err)
	s.merchantKey = mFull.ApiKey

	mCard, err := cardRepos.CardCommand.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID:       int(owner.UserID),
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(1, 0, 0),
		CVV:          "321",
		CardProvider: "mastercard",
	})
	s.Require().NoError(err)
	_, err = saldoRepos.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{
		CardNumber:   mCard.CardNumber,
		TotalBalance: 0,
	})
	s.Require().NoError(err)

	transactionHandler := handler.NewHandler(transactionService)
	server := grpc.NewServer()
	pb.RegisterTransactionCommandServiceServer(server, transactionHandler)
	pb.RegisterTransactionQueryServiceServer(server, transactionHandler)
	s.grpcServer = server

	lis, err := net.Listen("tcp", ":0")
	s.Require().NoError(err)
	go func() { _ = server.Serve(lis) }()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn
	s.txClient = pb.NewTransactionCommandServiceClient(conn)

	s.gql = tests.NewGraphQLClient(s.T(), &testhelper.ServiceConnections{TransactionClient: conn}, log, s.redisClient)
}

func (s *TransactionGraphqlTestSuite) TearDownSuite() {
	s.conn.Close()
	s.grpcServer.Stop()
	s.redisClient.Close()
	s.dbPool.Close()
	s.ts.Teardown()
}

// TestTransactionGraphqlReadLifecycle seeds a transaction through the gRPC
// command service (the create mutation is gated by a Kafka merchant check that
// is unavailable in the harness) and then exercises the full read/trash/
// restore/delete lifecycle through GraphQL.
func (s *TransactionGraphqlTestSuite) TestTransactionGraphqlReadLifecycle() {
	// Seed via gRPC, bypassing the gateway's Kafka-gated create mutation.
	seeded, err := s.txClient.CreateTransaction(context.Background(), &pb.CreateTransactionRequest{
		ApiKey:          s.merchantKey,
		CardNumber:      s.customerCard,
		Amount:          50000,
		PaymentMethod:   "visa",
		MerchantId:      int32(s.merchantID),
		TransactionTime: timestamppb.New(time.Now()),
		IdempotencyKey:  "graphql-seed-" + strconv.FormatInt(time.Now().UnixNano(), 10),
	})
	s.Require().NoError(err)
	s.Require().NotNil(seeded.GetData())
	s.transactionID = int(seeded.GetData().GetId())
	s.Require().NotZero(s.transactionID)

	// 1. FindById
	found := s.gql.Do(`
		query Find($input: FindByIdTransactionInput!) { findByIdTransaction(input: $input) { status data { id card_number } } }`,
		map[string]interface{}{"input": map[string]interface{}{"transaction_id": s.transactionID}})
	foundData := tests.Payload(s.T(), found, "findByIdTransaction")
	s.Equal(s.transactionID, tests.Int(s.T(), foundData, "id"))
	s.Equal(s.customerCard, tests.Str(s.T(), foundData, "card_number"))

	// 2. FindAll
	all := s.gql.Do(`query { findAllTransaction(input: {page: 1, page_size: 10}) { status data { id } } }`, nil)
	s.NotEmpty(tests.Items(s.T(), all, "findAllTransaction"))

	// 3. FindByMerchantId
	byMerchant := s.gql.Do(`
		query ByMerchant($input: FindTransactionByMerchantIdInput!) { findTransactionByMerchantId(input: $input) { status data { id merchant_id } } }`,
		map[string]interface{}{"input": map[string]interface{}{"merchant_id": s.merchantID}})
	s.NotEmpty(tests.Items(s.T(), byMerchant, "findTransactionByMerchantId"))

	// 4. FindByActive
	active := s.gql.Do(`query { findByActiveTransaction(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), active, "findByActiveTransaction"))

	// 5. Trash / findByTrashed / restore
	trashed := s.gql.Do(`
		mutation Trash($input: FindByIdTransactionInput!) { trashedTransaction(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"transaction_id": s.transactionID}})
	s.Equal("success", tests.Status(s.T(), trashed, "trashedTransaction"))

	trashedList := s.gql.Do(`query { findByTrashedTransaction(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), trashedList, "findByTrashedTransaction"))

	restored := s.gql.Do(`
		mutation Restore($input: FindByIdTransactionInput!) { restoreTransaction(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"transaction_id": s.transactionID}})
	s.Equal("success", tests.Status(s.T(), restored, "restoreTransaction"))

	// 6. DeletePermanent
	s.gql.Do(`
		mutation Trash($input: FindByIdTransactionInput!) { trashedTransaction(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"transaction_id": s.transactionID}})
	permanent := s.gql.Do(`
		mutation DeletePermanent($input: FindByIdTransactionInput!) { deleteTransactionPermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"transaction_id": s.transactionID}})
	s.Equal("success", tests.Status(s.T(), permanent, "deleteTransactionPermanent"))

	// 7. Bulk
	restoreAll := s.gql.Do(`mutation { restoreAllTransaction { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), restoreAll, "restoreAllTransaction"))
	deleteAll := s.gql.Do(`mutation { deleteAllTransactionPermanent { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), deleteAll, "deleteAllTransactionPermanent"))
}

// TestTransactionGraphqlCreateRequiresMerchantValidation asserts the create
// mutation's permission gate: the gateway validates the API key over Kafka, and
// with no Kafka broker configured the mutation must surface that as an error
// rather than silently creating a transaction.
func (s *TransactionGraphqlTestSuite) TestTransactionGraphqlCreateRequiresMerchantValidation() {
	_, errs := s.gql.DoRaw(`
		mutation Create($input: CreateTransactionInput!) {
			createTransaction(input: $input) { status data { id } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"api_key":          "unvalidated-api-key",
		"card_number":      s.customerCard,
		"amount":           10000,
		"payment_method":   "visa",
		"merchant_id":      s.merchantID,
		"transaction_time": time.Now().Format("2006-01-02"),
	}})
	s.Require().NotEmpty(errs, "createTransaction must be rejected when the merchant API key cannot be validated")
}

// TestTransactionGraphqlStats covers the transaction-scoped stats queries. The
// gateway resolves these from the stats-reader service; in the harness they are
// wired to the stub surface, so the assertion is on the response envelope.
func (s *TransactionGraphqlTestSuite) TestTransactionGraphqlStats() {
	year := time.Now().Year()
	month := int(time.Now().Month())
	card := s.customerCard

	cases := []struct {
		field string
		query string
		vars  map[string]interface{}
	}{
		{"findMonthlyTransactionAmountStats", `query($i: FindYearStatsInput!){ findMonthlyTransactionAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findYearlyTransactionAmountStats", `query($i: FindYearStatsInput!){ findYearlyTransactionAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findMonthlyTransactionAmountByCardNumberStats", `query($i: FindYearCardNumberStatsInput!){ findMonthlyTransactionAmountByCardNumberStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "card_number": card}}},
		{"findYearlyTransactionAmountByCardNumberStats", `query($i: FindYearCardNumberStatsInput!){ findYearlyTransactionAmountByCardNumberStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "card_number": card}}},

		{"findMonthlyTransactionMethodsStats", `query($i: FindYearStatsInput!){ findMonthlyTransactionMethodsStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findYearlyTransactionMethodsStats", `query($i: FindYearStatsInput!){ findYearlyTransactionMethodsStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},

		{"findMonthlyTransactionStatusSuccessStats", `query($i: FindMonthlyYearStatsInput!){ findMonthlyTransactionStatusSuccessStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "month": month}}},
		{"findYearlyTransactionStatusSuccessStats", `query($i: FindYearStatsInput!){ findYearlyTransactionStatusSuccessStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findMonthlyTransactionStatusFailedStats", `query($i: FindMonthlyYearStatsInput!){ findMonthlyTransactionStatusFailedStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "month": month}}},
		{"findYearlyTransactionStatusFailedStats", `query($i: FindYearStatsInput!){ findYearlyTransactionStatusFailedStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
	}
	for _, tc := range cases {
		data := s.gql.Do(tc.query, tc.vars)
		s.Equal("success", tests.Status(s.T(), data, tc.field), tc.field)
	}
}

func TestTransactionGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransactionGraphqlTestSuite))
}
