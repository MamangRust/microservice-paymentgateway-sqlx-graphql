package merchant_test

import (
	"context"
	"net"
	"testing"
	"time"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/testhelper"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/service"
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
)

type MerchantGraphqlTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	dbPool      *sqlx.DB
	redisClient *redis.Client
	grpcServer  *grpc.Server
	conn        *grpc.ClientConn
	gql         *tests.GraphQLClient
	userID      int
	merchantID  int
	apiKey      string
}

func (s *MerchantGraphqlTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card", "merchant"))

	pool, err := sqlx.ConnectContext(s.ts.Ctx, "pgx", s.ts.DBURL)
	s.Require().NoError(err)
	s.dbPool = pool

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	merchantRepos := repository.NewRepositories(pool, nil, nil)
	userRepos := user_repo.NewRepositories(pool, nil, nil)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	merchantService := service.NewService(&service.Deps{
		Kafka:        nil,
		Repositories: merchantRepos,
		UserAdapter:  s.ts.UserAdapter,
		Logger:       log,
		Cache:        cacheStore,
	})

	user, err := userRepos.UserCommand().CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Graphql",
		LastName:  "Merchant",
		Email:     "graphql.merchant@example.com",
		Password:  "password123",
	})
	s.Require().NoError(err)
	s.userID = int(user.UserID)

	merchantHandler := handler.NewHandler(merchantService)
	server := grpc.NewServer()
	pb.RegisterMerchantCommandServiceServer(server, merchantHandler)
	pb.RegisterMerchantQueryServiceServer(server, merchantHandler)
	s.grpcServer = server

	lis, err := net.Listen("tcp", ":0")
	s.Require().NoError(err)
	go func() { _ = server.Serve(lis) }()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.gql = tests.NewGraphQLClient(s.T(), &testhelper.ServiceConnections{MerchantClient: conn}, log, s.redisClient)
}

func (s *MerchantGraphqlTestSuite) TearDownSuite() {
	s.conn.Close()
	s.grpcServer.Stop()
	s.redisClient.Close()
	s.dbPool.Close()
	s.ts.Teardown()
}

func (s *MerchantGraphqlTestSuite) TestMerchantGraphqlLifecycle() {
	// 1. Create
	create := s.gql.Do(`
		mutation Create($input: CreateMerchantInput!) {
			createMerchant(input: $input) { status message data { id name api_key user_id } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"name":    "Graphql Merchant",
		"user_id": s.userID,
	}})
	created := tests.Payload(s.T(), create, "createMerchant")
	s.merchantID = tests.Int(s.T(), created, "id")
	s.apiKey = tests.Str(s.T(), created, "api_key")
	s.Require().NotZero(s.merchantID)
	s.Equal("Graphql Merchant", tests.Str(s.T(), created, "name"))

	// 2. FindById
	found := s.gql.Do(`
		query Find($input: FindByIdMerchantInput!) { findByIdMerchant(input: $input) { status data { id } } }`,
		map[string]interface{}{"input": map[string]interface{}{"merchant_id": s.merchantID}})
	s.Equal(s.merchantID, tests.Int(s.T(), tests.Payload(s.T(), found, "findByIdMerchant"), "id"))

	// 3. FindByApiKey
	byKey := s.gql.Do(`
		query ByKey($input: FindByApiKeyInput!) { findByApiKey(input: $input) { status data { id } } }`,
		map[string]interface{}{"input": map[string]interface{}{"api_key": s.apiKey}})
	s.Equal(s.merchantID, tests.Int(s.T(), tests.Payload(s.T(), byKey, "findByApiKey"), "id"))

	// 4. FindByMerchantUserId (list)
	byUser := s.gql.Do(`
		query ByUser($input: FindByMerchantUserIdInput!) { findByMerchantUserId(input: $input) { status data { id } } }`,
		map[string]interface{}{"input": map[string]interface{}{"user_id": s.userID}})
	s.NotEmpty(tests.Items(s.T(), byUser, "findByMerchantUserId"))

	// 5. FindAll
	all := s.gql.Do(`query { findAllMerchant(input: {page: 1, page_size: 10}) { status data { id } } }`, nil)
	s.NotEmpty(tests.Items(s.T(), all, "findAllMerchant"))

	// 6. Update
	update := s.gql.Do(`
		mutation Update($input: UpdateMerchantInput!) {
			updateMerchant(input: $input) { status data { id name status } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"merchant_id": s.merchantID,
		"name":        "Graphql Merchant Updated",
		"user_id":     s.userID,
		"status":      "active",
	}})
	updated := tests.Payload(s.T(), update, "updateMerchant")
	s.Equal("Graphql Merchant Updated", tests.Str(s.T(), updated, "name"))
	s.Equal("active", tests.Str(s.T(), updated, "status"))

	// NOTE: updateMerchantStatus is exposed by the GraphQL schema but the
	// merchant gRPC service does not implement it (returns Unimplemented), so
	// it is intentionally not exercised here. Status is covered via updateMerchant.

	// 7. Trash / restore
	trashed := s.gql.Do(`
		mutation Trash($input: FindByIdMerchantInput!) { trashedMerchant(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"merchant_id": s.merchantID}})
	s.Equal("success", tests.Status(s.T(), trashed, "trashedMerchant"))

	trashedList := s.gql.Do(`query { findByTrashed(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), trashedList, "findByTrashed"))

	restored := s.gql.Do(`
		mutation Restore($input: FindByIdMerchantInput!) { restoreMerchant(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"merchant_id": s.merchantID}})
	s.Equal("success", tests.Status(s.T(), restored, "restoreMerchant"))

	// 9. DeletePermanent
	s.gql.Do(`
		mutation Trash($input: FindByIdMerchantInput!) { trashedMerchant(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"merchant_id": s.merchantID}})
	permanent := s.gql.Do(`
		mutation DeletePermanent($input: FindByIdMerchantInput!) { deleteMerchantPermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"merchant_id": s.merchantID}})
	s.Equal("success", tests.Status(s.T(), permanent, "deleteMerchantPermanent"))

	// 10. Bulk
	restoreAll := s.gql.Do(`mutation { restoreAllMerchant { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), restoreAll, "restoreAllMerchant"))
	deleteAll := s.gql.Do(`mutation { deleteAllMerchantPermanent { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), deleteAll, "deleteAllMerchantPermanent"))
}

// TestMerchantGraphqlStats covers the merchant-scoped stats queries. The gateway
// resolves these from the stats-reader service; in the harness they are wired to
// the stub surface, so the assertion is on the response envelope.
func (s *MerchantGraphqlTestSuite) TestMerchantGraphqlStats() {
	year := time.Now().Year()
	merchantID := s.merchantID

	// The merchant seeded by the lifecycle test is deleted by the time stats
	// run only if tests share state; keep a dedicated id for the by-id inputs.
	if merchantID == 0 {
		merchantID = 1
	}

	cases := []struct {
		field string
		query string
		vars  map[string]interface{}
	}{
		{"findMonthlyMerchantAmountStats", `query($i: FindYearMerchantStatsInput!){ findMonthlyMerchantAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findYearlyMerchantAmountStats", `query($i: FindYearMerchantStatsInput!){ findYearlyMerchantAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findMonthlyMerchantMethodStats", `query($i: FindYearMerchantStatsInput!){ findMonthlyMerchantMethodStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findYearlyMerchantMethodStats", `query($i: FindYearMerchantStatsInput!){ findYearlyMerchantMethodStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findMonthlyMerchantTotalAmountStats", `query($i: FindYearMerchantStatsInput!){ findMonthlyMerchantTotalAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findYearlyMerchantTotalAmountStats", `query($i: FindYearMerchantStatsInput!){ findYearlyMerchantTotalAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},

		{"findMonthlyMerchantAmountByIdStats", `query($i: FindYearMerchantByIdStatsInput!){ findMonthlyMerchantAmountByIdStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "merchant_id": merchantID}}},
		{"findYearlyMerchantAmountByIdStats", `query($i: FindYearMerchantByIdStatsInput!){ findYearlyMerchantAmountByIdStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "merchant_id": merchantID}}},
		{"findMonthlyMerchantMethodByIdStats", `query($i: FindYearMerchantByIdStatsInput!){ findMonthlyMerchantMethodByIdStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "merchant_id": merchantID}}},
		{"findYearlyMerchantMethodByIdStats", `query($i: FindYearMerchantByIdStatsInput!){ findYearlyMerchantMethodByIdStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "merchant_id": merchantID}}},
		{"findMonthlyMerchantTotalAmountByIdStats", `query($i: FindYearMerchantByIdStatsInput!){ findMonthlyMerchantTotalAmountByIdStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "merchant_id": merchantID}}},
		{"findYearlyMerchantTotalAmountByIdStats", `query($i: FindYearMerchantByIdStatsInput!){ findYearlyMerchantTotalAmountByIdStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "merchant_id": merchantID}}},

		{"findMonthlyMerchantAmountByApikeyStats", `query($i: FindYearMerchantByApikeyStatsInput!){ findMonthlyMerchantAmountByApikeyStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "api_key": "api-key-test"}}},
		{"findYearlyMerchantAmountByApikeyStats", `query($i: FindYearMerchantByApikeyStatsInput!){ findYearlyMerchantAmountByApikeyStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "api_key": "api-key-test"}}},
		{"findMonthlyMerchantMethodByApikeyStats", `query($i: FindYearMerchantByApikeyStatsInput!){ findMonthlyMerchantMethodByApikeyStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "api_key": "api-key-test"}}},
		{"findYearlyMerchantMethodByApikeyStats", `query($i: FindYearMerchantByApikeyStatsInput!){ findYearlyMerchantMethodByApikeyStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "api_key": "api-key-test"}}},
		{"findMonthlyMerchantTotalAmountByApikeyStats", `query($i: FindYearMerchantByApikeyStatsInput!){ findMonthlyMerchantTotalAmountByApikeyStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "api_key": "api-key-test"}}},
		{"findYearlyMerchantTotalAmountByApikeyStats", `query($i: FindYearMerchantByApikeyStatsInput!){ findYearlyMerchantTotalAmountByApikeyStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "api_key": "api-key-test"}}},

		{"findMerchantTransactionsStats", `query($i: FindMerchantTransactionsInput!){ findMerchantTransactionsStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"page": 1, "page_size": 10, "merchant_id": merchantID}}},
		{"findMerchantTransactionsByApikeyStats", `query($i: FindMerchantTransactionsByApikeyInput!){ findMerchantTransactionsByApikeyStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"page": 1, "page_size": 10, "api_key": "api-key-test"}}},
	}
	for _, tc := range cases {
		data := s.gql.Do(tc.query, tc.vars)
		s.Equal("success", tests.Status(s.T(), data, tc.field), tc.field)
	}
}

func TestMerchantGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantGraphqlTestSuite))
}
