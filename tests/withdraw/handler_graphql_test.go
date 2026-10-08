package withdraw_test

import (
	"context"
	"net"
	"testing"
	"time"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/withdraw"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/testhelper"
	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	saldo_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	user_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw/service"
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

type WithdrawGraphqlTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	dbPool      *sqlx.DB
	redisClient *redis.Client
	grpcServer  *grpc.Server
	conn        *grpc.ClientConn
	gql         *tests.GraphQLClient
	cardNumber  string
	withdrawID  int
}

func (s *WithdrawGraphqlTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card", "saldo", "withdraw"))

	pool, err := sqlx.ConnectContext(s.ts.Ctx, "pgx", s.ts.DBURL)
	s.Require().NoError(err)
	s.dbPool = pool

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	userRepos := user_repo.NewRepositories(pool, nil, nil)
	cardRepos := card_repo.NewRepositories(pool, nil)
	saldoRepos := saldo_repo.NewRepositories(pool, nil, nil)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	withdrawService := service.NewService(&service.Deps{
		Kafka:            nil,
		Repositories:     repository.NewRepositories(&repository.Deps{DB: pool}),
		CardAdapter:      s.ts.CardAdapter,
		SaldoAdapter:     s.ts.SaldoAdapter,
		Logger:           log,
		Cache:            cacheStore,
		AISecurityClient: nil,
	})

	user, err := userRepos.UserCommand().CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Graphql",
		LastName:  "Withdraw",
		Email:     "graphql.withdraw@example.com",
		Password:  "password123",
	})
	s.Require().NoError(err)

	card, err := cardRepos.CardCommand.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID:       int(user.UserID),
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(1, 0, 0),
		CVV:          "999",
		CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.cardNumber = card.CardNumber

	_, err = saldoRepos.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{
		CardNumber:   s.cardNumber,
		TotalBalance: 1000000,
	})
	s.Require().NoError(err)

	withdrawHandler := handler.NewHandler(withdrawService)
	server := grpc.NewServer()
	pb.RegisterWithdrawCommandServiceServer(server, withdrawHandler)
	pb.RegisterWithdrawQueryServiceServer(server, withdrawHandler)
	s.grpcServer = server

	lis, err := net.Listen("tcp", ":0")
	s.Require().NoError(err)
	go func() { _ = server.Serve(lis) }()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.gql = tests.NewGraphQLClient(s.T(), &testhelper.ServiceConnections{WithdrawClient: conn}, log, s.redisClient)
}

func (s *WithdrawGraphqlTestSuite) TearDownSuite() {
	s.conn.Close()
	s.grpcServer.Stop()
	s.redisClient.Close()
	s.dbPool.Close()
	s.ts.Teardown()
}

func (s *WithdrawGraphqlTestSuite) TestWithdrawGraphqlLifecycle() {
	withdrawTime := time.Now().Format("2006-01-02")

	// 1. Create
	create := s.gql.Do(`
		mutation Create($input: CreateWithdrawInput!) {
			createWithdraw(input: $input) { status message data { withdraw_id card_number withdraw_amount } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"card_number":     s.cardNumber,
		"withdraw_amount": 100000,
		"withdraw_time":   withdrawTime,
	}})
	created := tests.Payload(s.T(), create, "createWithdraw")
	s.withdrawID = tests.Int(s.T(), created, "withdraw_id")
	s.Require().NotZero(s.withdrawID)
	s.Equal(s.cardNumber, tests.Str(s.T(), created, "card_number"))

	// 2. FindById
	found := s.gql.Do(`
		query Find($input: FindByIdWithdrawInput!) { findByIdWithdraw(input: $input) { status data { withdraw_id } } }`,
		map[string]interface{}{"input": map[string]interface{}{"withdraw_id": s.withdrawID}})
	s.Equal(s.withdrawID, tests.Int(s.T(), tests.Payload(s.T(), found, "findByIdWithdraw"), "withdraw_id"))

	// 3. FindAll
	all := s.gql.Do(`query { findAllWithdraw(input: {page: 1, page_size: 10}) { status data { withdraw_id } } }`, nil)
	s.NotEmpty(tests.Items(s.T(), all, "findAllWithdraw"))

	// 4. Update
	update := s.gql.Do(`
		mutation Update($input: UpdateWithdrawInput!) {
			updateWithdraw(input: $input) { status data { withdraw_id withdraw_amount } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"withdraw_id":     s.withdrawID,
		"card_number":     s.cardNumber,
		"withdraw_amount": 150000,
		"withdraw_time":   withdrawTime,
	}})
	s.Equal(150000, tests.Int(s.T(), tests.Payload(s.T(), update, "updateWithdraw"), "withdraw_amount"))

	// 5. Trash / restore
	trashed := s.gql.Do(`
		mutation Trash($input: FindByIdWithdrawInput!) { trashedWithdraw(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"withdraw_id": s.withdrawID}})
	s.Equal("success", tests.Status(s.T(), trashed, "trashedWithdraw"))

	trashedList := s.gql.Do(`query { findByTrashedWithdraw(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), trashedList, "findByTrashedWithdraw"))

	restored := s.gql.Do(`
		mutation Restore($input: FindByIdWithdrawInput!) { restoreWithdraw(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"withdraw_id": s.withdrawID}})
	s.Equal("success", tests.Status(s.T(), restored, "restoreWithdraw"))

	// 6. DeletePermanent
	s.gql.Do(`
		mutation Trash($input: FindByIdWithdrawInput!) { trashedWithdraw(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"withdraw_id": s.withdrawID}})
	permanent := s.gql.Do(`
		mutation DeletePermanent($input: FindByIdWithdrawInput!) { deleteWithdrawPermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"withdraw_id": s.withdrawID}})
	s.Equal("success", tests.Status(s.T(), permanent, "deleteWithdrawPermanent"))

	// 7. Bulk
	restoreAll := s.gql.Do(`mutation { restoreAllWithdraw { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), restoreAll, "restoreAllWithdraw"))
	deleteAll := s.gql.Do(`mutation { deleteAllWithdrawPermanent { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), deleteAll, "deleteAllWithdrawPermanent"))
}

// TestWithdrawGraphqlStats covers the withdraw-scoped stats queries. The gateway
// resolves these from the stats-reader service; in the harness they are wired to
// the stub surface, so the assertion is on the response envelope.
func (s *WithdrawGraphqlTestSuite) TestWithdrawGraphqlStats() {
	year := time.Now().Year()
	month := int(time.Now().Month())

	cases := []struct {
		field string
		query string
		vars  map[string]interface{}
	}{
		{"findMonthlyWithdrawAmountStats", `query($i: FindYearStatsInput!){ findMonthlyWithdrawAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findYearlyWithdrawAmountStats", `query($i: FindYearStatsInput!){ findYearlyWithdrawAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findMonthlyWithdrawAmountByCardNumberStats", `query($i: FindYearCardNumberStatsInput!){ findMonthlyWithdrawAmountByCardNumberStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "card_number": s.cardNumber}}},
		{"findYearlyWithdrawAmountByCardNumberStats", `query($i: FindYearCardNumberStatsInput!){ findYearlyWithdrawAmountByCardNumberStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "card_number": s.cardNumber}}},

		{"findMonthlyWithdrawStatusSuccessStats", `query($i: FindMonthlyYearStatsInput!){ findMonthlyWithdrawStatusSuccessStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "month": month}}},
		{"findYearlyWithdrawStatusSuccessStats", `query($i: FindYearStatsInput!){ findYearlyWithdrawStatusSuccessStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findMonthlyWithdrawStatusFailedStats", `query($i: FindMonthlyYearStatsInput!){ findMonthlyWithdrawStatusFailedStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "month": month}}},
		{"findYearlyWithdrawStatusFailedStats", `query($i: FindYearStatsInput!){ findYearlyWithdrawStatusFailedStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
	}
	for _, tc := range cases {
		data := s.gql.Do(tc.query, tc.vars)
		s.Equal("success", tests.Status(s.T(), data, tc.field), tc.field)
	}
}

func TestWithdrawGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(WithdrawGraphqlTestSuite))
}
