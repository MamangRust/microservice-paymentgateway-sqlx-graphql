package card_test

import (
	"context"
	"net"
	"testing"
	"time"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/testhelper"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/card/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/card/service"
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

type CardGraphqlTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	dbPool      *sqlx.DB
	redisClient *redis.Client
	grpcServer  *grpc.Server
	conn        *grpc.ClientConn
	gql         *tests.GraphQLClient
	userID      int
	cardID      int
	cardNumber  string
}

func (s *CardGraphqlTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card"))

	pool, err := sqlx.ConnectContext(s.ts.Ctx, "pgx", s.ts.DBURL)
	s.Require().NoError(err)
	s.dbPool = pool

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	cardRepos := repository.NewRepositories(pool, nil)
	userRepos := user_repo.NewRepositories(pool, nil, nil)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	cardService := service.NewService(&service.Deps{
		Repositories: cardRepos,
		UserAdapter:  s.ts.UserAdapter,
		Logger:       log,
		Cache:        cacheStore,
		Kafka:        nil,
	})

	// A card must belong to a real user, so seed one up front.
	user, err := userRepos.UserCommand().CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Graphql",
		LastName:  "Card",
		Email:     "graphql.card@example.com",
		Password:  "password123",
	})
	s.Require().NoError(err)
	s.userID = int(user.UserID)

	cardHandler := handler.NewHandler(cardService)
	server := grpc.NewServer()
	pb.RegisterCardQueryServiceServer(server, cardHandler)
	pb.RegisterCardCommandServiceServer(server, cardHandler)
	s.grpcServer = server

	lis, err := net.Listen("tcp", ":0")
	s.Require().NoError(err)
	go func() { _ = server.Serve(lis) }()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.gql = tests.NewGraphQLClient(s.T(), &testhelper.ServiceConnections{CardClient: conn}, log, s.redisClient)
}

func (s *CardGraphqlTestSuite) TearDownSuite() {
	s.conn.Close()
	s.grpcServer.Stop()
	s.redisClient.Close()
	s.dbPool.Close()
	s.ts.Teardown()
}

func (s *CardGraphqlTestSuite) TestCardGraphqlLifecycle() {
	expireDate := time.Now().AddDate(5, 0, 0).Format("2006-01-02")

	// 1. Create
	create := s.gql.Do(`
		mutation Create($input: CreateCardInput!) {
			createCard(input: $input) { status message data { id user_id card_number card_type card_provider } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"user_id":       s.userID,
		"card_type":     "debit",
		"expire_date":   expireDate,
		"cvv":           "123",
		"card_provider": "visa",
	}})
	created := tests.Payload(s.T(), create, "createCard")
	s.cardID = tests.Int(s.T(), created, "id")
	s.cardNumber = tests.Str(s.T(), created, "card_number")
	s.Require().NotZero(s.cardID)
	s.Equal(s.userID, tests.Int(s.T(), created, "user_id"))

	// 2. FindById
	found := s.gql.Do(`
		query Find($input: FindByIdCardInput!) { findByIdCard(input: $input) { status data { id } } }`,
		map[string]interface{}{"input": map[string]interface{}{"card_id": s.cardID}})
	s.Equal(s.cardID, tests.Int(s.T(), tests.Payload(s.T(), found, "findByIdCard"), "id"))

	// 3. FindByUserId
	byUser := s.gql.Do(`
		query ByUser($input: FindByUserIdCardInput!) { findByUserIdCard(input: $input) { status data { user_id } } }`,
		map[string]interface{}{"input": map[string]interface{}{"user_id": s.userID}})
	s.Equal(s.userID, tests.Int(s.T(), tests.Payload(s.T(), byUser, "findByUserIdCard"), "user_id"))

	// 4. FindByCardNumber
	byNumber := s.gql.Do(`
		query ByNumber($input: FindByCardNumberInput!) { findByCardNumber(input: $input) { status data { card_number } } }`,
		map[string]interface{}{"input": map[string]interface{}{"card_number": s.cardNumber}})
	s.Equal(s.cardNumber, tests.Str(s.T(), tests.Payload(s.T(), byNumber, "findByCardNumber"), "card_number"))

	// 5. FindAll
	all := s.gql.Do(`query { findAllCard(input: {page: 1, page_size: 10}) { status data { id } } }`, nil)
	s.NotEmpty(tests.Items(s.T(), all, "findAllCard"))

	// 6. Update
	update := s.gql.Do(`
		mutation Update($input: UpdateCardInput!) {
			updateCard(input: $input) { status data { id card_type card_provider } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"card_id":       s.cardID,
		"user_id":       s.userID,
		"card_type":     "credit",
		"expire_date":   time.Now().AddDate(6, 0, 0).Format("2006-01-02"),
		"cvv":           "456",
		"card_provider": "mastercard",
	}})
	updated := tests.Payload(s.T(), update, "updateCard")
	s.Equal("credit", tests.Str(s.T(), updated, "card_type"))
	s.Equal("mastercard", tests.Str(s.T(), updated, "card_provider"))

	// 7. Trash / restore
	trashed := s.gql.Do(`
		mutation Trash($input: FindByIdCardInput!) { trashedCard(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"card_id": s.cardID}})
	s.Equal("success", tests.Status(s.T(), trashed, "trashedCard"))

	trashedList := s.gql.Do(`query { findByTrashedCard(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), trashedList, "findByTrashedCard"))

	restored := s.gql.Do(`
		mutation Restore($input: FindByIdCardInput!) { restoreCard(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"card_id": s.cardID}})
	s.Equal("success", tests.Status(s.T(), restored, "restoreCard"))

	// 8. DeletePermanent
	s.gql.Do(`
		mutation Trash($input: FindByIdCardInput!) { trashedCard(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"card_id": s.cardID}})
	permanent := s.gql.Do(`
		mutation DeletePermanent($input: FindByIdCardInput!) { deleteCardPermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"card_id": s.cardID}})
	s.Equal("success", tests.Status(s.T(), permanent, "deleteCardPermanent"))

	// 9. Bulk
	restoreAll := s.gql.Do(`mutation { restoreAllCard { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), restoreAll, "restoreAllCard"))
	deleteAll := s.gql.Do(`mutation { deleteAllCardPermanent { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), deleteAll, "deleteAllCardPermanent"))
}

// TestCardGraphqlStats covers the card-scoped stats queries. The gateway
// resolves these from the stats-reader service; in the harness they are wired
// to the same stub surface as the rest of the gateway, so the assertion is on
// the response envelope rather than on ClickHouse-backed numbers.
func (s *CardGraphqlTestSuite) TestCardGraphqlStats() {
	year := time.Now().Year()
	card := s.cardNumber

	cases := []struct {
		field string
		query string
		vars  map[string]interface{}
	}{
		{"findMonthlyBalanceStats", `query($i: FindYearStatsInput!){ findMonthlyBalanceStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findYearlyBalanceStats", `query($i: FindYearStatsInput!){ findYearlyBalanceStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findMonthlyBalanceByCardNumberStats", `query($i: FindYearCardNumberStatsInput!){ findMonthlyBalanceByCardNumberStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "card_number": card}}},
		{"findYearlyBalanceByCardNumberStats", `query($i: FindYearCardNumberStatsInput!){ findYearlyBalanceByCardNumberStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "card_number": card}}},

		{"findMonthlyTopupAmountStats", `query($i: FindYearStatsInput!){ findMonthlyTopupAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findYearlyTopupAmountStats", `query($i: FindYearStatsInput!){ findYearlyTopupAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findMonthlyTopupAmountByCardNumberStats", `query($i: FindYearCardNumberStatsInput!){ findMonthlyTopupAmountByCardNumberStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "card_number": card}}},
		{"findYearlyTopupAmountByCardNumberStats", `query($i: FindYearCardNumberStatsInput!){ findYearlyTopupAmountByCardNumberStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "card_number": card}}},

		{"findMonthlyWithdrawAmountStats", `query($i: FindYearStatsInput!){ findMonthlyWithdrawAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findYearlyWithdrawAmountStats", `query($i: FindYearStatsInput!){ findYearlyWithdrawAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findMonthlyWithdrawAmountByCardNumberStats", `query($i: FindYearCardNumberStatsInput!){ findMonthlyWithdrawAmountByCardNumberStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "card_number": card}}},
		{"findYearlyWithdrawAmountByCardNumberStats", `query($i: FindYearCardNumberStatsInput!){ findYearlyWithdrawAmountByCardNumberStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "card_number": card}}},

		{"findMonthlyTransactionAmountStats", `query($i: FindYearStatsInput!){ findMonthlyTransactionAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findYearlyTransactionAmountStats", `query($i: FindYearStatsInput!){ findYearlyTransactionAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findMonthlyTransactionAmountByCardNumberStats", `query($i: FindYearCardNumberStatsInput!){ findMonthlyTransactionAmountByCardNumberStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "card_number": card}}},
		{"findYearlyTransactionAmountByCardNumberStats", `query($i: FindYearCardNumberStatsInput!){ findYearlyTransactionAmountByCardNumberStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "card_number": card}}},

		{"findMonthlyTransferAmountStats", `query($i: FindYearStatsInput!){ findMonthlyTransferAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findYearlyTransferAmountStats", `query($i: FindYearStatsInput!){ findYearlyTransferAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findMonthlyTransferSenderAmountStats", `query($i: FindYearCardNumberStatsInput!){ findMonthlyTransferSenderAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "card_number": card}}},
		{"findYearlyTransferSenderAmountStats", `query($i: FindYearCardNumberStatsInput!){ findYearlyTransferSenderAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "card_number": card}}},
		{"findMonthlyTransferReceiverAmountStats", `query($i: FindYearCardNumberStatsInput!){ findMonthlyTransferReceiverAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "card_number": card}}},
		{"findYearlyTransferReceiverAmountStats", `query($i: FindYearCardNumberStatsInput!){ findYearlyTransferReceiverAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "card_number": card}}},

		{"dashboardCardStats", `query { dashboardCardStats { status } }`, nil},
		{"dashboardCardByNumberStats", `query($i: FindByCardNumberInput!){ dashboardCardByNumberStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"card_number": card}}},
	}
	for _, tc := range cases {
		data := s.gql.Do(tc.query, tc.vars)
		s.Equal("success", tests.Status(s.T(), data, tc.field), tc.field)
	}
}

func TestCardGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CardGraphqlTestSuite))
}
