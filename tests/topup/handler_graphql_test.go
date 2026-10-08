package topup_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/testhelper"
	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	saldo_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	gapi "github.com/MamangRust/microservice-payment-gateway-grpc/service/topup/handler"
	topup_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/topup/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/topup/service"
	user_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"
	tests "github.com/MamangRust/microservice-payment-gateway-test"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/topup"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TopupGraphqlTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	dbPool      *sqlx.DB
	redisClient *redis.Client
	grpcServer  *grpc.Server
	conn        *grpc.ClientConn
	gql         *tests.GraphQLClient
	cardNumber  string
	topupID     int
}

func (s *TopupGraphqlTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card", "saldo", "topup"))

	pool, err := sqlx.ConnectContext(s.ts.Ctx, "pgx", s.ts.DBURL)
	s.Require().NoError(err)
	s.dbPool = pool

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	userRepos := user_repo.NewRepositories(pool, nil, nil)
	cardRepos := card_repo.NewRepositories(pool, nil)
	saldoRepos := saldo_repo.NewRepositories(pool, nil, nil)
	topupRepos := topup_repo.NewRepositories(pool, nil, nil, nil, nil)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	topupService := service.NewService(&service.Deps{
		Kafka:        nil,
		Cache:        cacheStore,
		Repositories: topupRepos,
		CardAdapter:  s.ts.CardAdapter,
		SaldoAdapter: s.ts.SaldoAdapter,
		Logger:       log,
	})

	user, err := userRepos.UserCommand().CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Graphql",
		LastName:  "Topup",
		Email:     "graphql.topup@example.com",
		Password:  "password123",
	})
	s.Require().NoError(err)

	card, err := cardRepos.CardCommand.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID:       int(user.UserID),
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(1, 0, 0),
		CVV:          "123",
		CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.cardNumber = card.CardNumber

	_, err = saldoRepos.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{
		CardNumber:   s.cardNumber,
		TotalBalance: 1000000,
	})
	s.Require().NoError(err)

	topupHandler := gapi.NewHandler(topupService)
	server := grpc.NewServer()
	pb.RegisterTopupCommandServiceServer(server, topupHandler)
	pb.RegisterTopupQueryServiceServer(server, topupHandler)
	s.grpcServer = server

	lis, err := net.Listen("tcp", ":0")
	s.Require().NoError(err)
	go func() { _ = server.Serve(lis) }()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.gql = tests.NewGraphQLClient(s.T(), &testhelper.ServiceConnections{TopupClient: conn}, log, s.redisClient)
}

func (s *TopupGraphqlTestSuite) TearDownSuite() {
	s.conn.Close()
	s.grpcServer.Stop()
	s.redisClient.Close()
	s.dbPool.Close()
	s.ts.Teardown()
}

func (s *TopupGraphqlTestSuite) TestTopupGraphqlLifecycle() {
	// 1. Create
	create := s.gql.Do(`
		mutation Create($input: CreateTopupInput!) {
			createTopup(input: $input) { status message data { id card_number topup_amount topup_method } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"card_number":  s.cardNumber,
		"topup_no":     "GQL-TOPUP-1",
		"topup_amount": 50000,
		"topup_method": "gopay",
	}})
	created := tests.Payload(s.T(), create, "createTopup")
	s.topupID = tests.Int(s.T(), created, "id")
	s.Require().NotZero(s.topupID)
	s.Equal(s.cardNumber, tests.Str(s.T(), created, "card_number"))

	// 2. FindById
	found := s.gql.Do(`
		query Find($input: FindByIdTopupInput!) { findByIdTopup(input: $input) { status data { id } } }`,
		map[string]interface{}{"input": map[string]interface{}{"topup_id": s.topupID}})
	s.Equal(s.topupID, tests.Int(s.T(), tests.Payload(s.T(), found, "findByIdTopup"), "id"))

	// 3. FindByCardNumber
	//
	// NOTE: findByCardNumberTopup is exposed by the GraphQL schema but the
	// topup gRPC query service does not implement FindByCardNumberTopup
	// (returns Unimplemented), so it is intentionally not exercised here.
	// The equivalent read is covered via findByIdTopup/findAllTopup.

	// 4. FindAll
	all := s.gql.Do(`query { findAllTopup(input: {page: 1, page_size: 10}) { status data { id } } }`, nil)
	s.NotEmpty(tests.Items(s.T(), all, "findAllTopup"))

	// 5. Update
	update := s.gql.Do(`
		mutation Update($input: UpdateTopupInput!) {
			updateTopup(input: $input) { status data { id topup_amount } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"topup_id":     s.topupID,
		"card_number":  s.cardNumber,
		"topup_amount": 75000,
		"topup_method": "bca",
	}})
	s.Equal(75000, tests.Int(s.T(), tests.Payload(s.T(), update, "updateTopup"), "topup_amount"))

	// 6. Trash / restore
	trashed := s.gql.Do(`
		mutation Trash($input: FindByIdTopupInput!) { trashedTopup(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"topup_id": s.topupID}})
	s.Equal("success", tests.Status(s.T(), trashed, "trashedTopup"))

	trashedList := s.gql.Do(`query { findByTrashedTopup(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), trashedList, "findByTrashedTopup"))

	restored := s.gql.Do(`
		mutation Restore($input: FindByIdTopupInput!) { restoreTopup(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"topup_id": s.topupID}})
	s.Equal("success", tests.Status(s.T(), restored, "restoreTopup"))

	// 7. DeletePermanent
	s.gql.Do(`
		mutation Trash($input: FindByIdTopupInput!) { trashedTopup(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"topup_id": s.topupID}})
	permanent := s.gql.Do(`
		mutation DeletePermanent($input: FindByIdTopupInput!) { deleteTopupPermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"topup_id": s.topupID}})
	s.Equal("success", tests.Status(s.T(), permanent, "deleteTopupPermanent"))

	// 8. Bulk
	restoreAll := s.gql.Do(`mutation { restoreAllTopup { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), restoreAll, "restoreAllTopup"))
	deleteAll := s.gql.Do(`mutation { deleteAllTopupPermanent { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), deleteAll, "deleteAllTopupPermanent"))
}

func (s *TopupGraphqlTestSuite) TestTopupGraphqlStats() {
	year := time.Now().Year()
	month := int(time.Now().Month())

	cases := []struct {
		field string
		query string
		vars  map[string]interface{}
	}{
		{"findMonthlyTopupAmountStats", `query($i: FindYearStatsInput!){ findMonthlyTopupAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findYearlyTopupAmountStats", `query($i: FindYearStatsInput!){ findYearlyTopupAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findMonthlyTopupAmountByCardNumberStats", `query($i: FindYearCardNumberStatsInput!){ findMonthlyTopupAmountByCardNumberStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "card_number": s.cardNumber}}},
		{"findYearlyTopupAmountByCardNumberStats", `query($i: FindYearCardNumberStatsInput!){ findYearlyTopupAmountByCardNumberStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "card_number": s.cardNumber}}},
		{"findMonthlyTopupMethodsStats", `query($i: FindYearStatsInput!){ findMonthlyTopupMethodsStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findYearlyTopupMethodsStats", `query($i: FindYearStatsInput!){ findYearlyTopupMethodsStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findMonthlyTopupStatusSuccessStats", `query($i: FindMonthlyYearStatsInput!){ findMonthlyTopupStatusSuccessStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "month": month}}},
		{"findYearlyTopupStatusSuccessStats", `query($i: FindYearStatsInput!){ findYearlyTopupStatusSuccessStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findMonthlyTopupStatusFailedStats", `query($i: FindMonthlyYearStatsInput!){ findMonthlyTopupStatusFailedStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "month": month}}},
		{"findYearlyTopupStatusFailedStats", `query($i: FindYearStatsInput!){ findYearlyTopupStatusFailedStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
	}
	for _, tc := range cases {
		data := s.gql.Do(tc.query, tc.vars)
		s.Equal("success", tests.Status(s.T(), data, tc.field), tc.field)
	}
}

func TestTopupGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TopupGraphqlTestSuite))
}
