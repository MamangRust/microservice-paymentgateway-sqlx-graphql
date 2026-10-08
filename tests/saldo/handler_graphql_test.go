package saldo_test

import (
	"context"
	"net"
	"testing"
	"time"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/testhelper"
	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	gapi "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/handler"
	saldo_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/service"
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

type SaldoGraphqlTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	dbPool      *sqlx.DB
	redisClient *redis.Client
	grpcServer  *grpc.Server
	conn        *grpc.ClientConn
	gql         *tests.GraphQLClient
	cardNumber  string
	saldoID     int
}

func (s *SaldoGraphqlTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card", "saldo"))

	pool, err := sqlx.ConnectContext(s.ts.Ctx, "pgx", s.ts.DBURL)
	s.Require().NoError(err)
	s.dbPool = pool

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	saldoRepos := saldo_repo.NewRepositories(pool, nil, nil)
	userRepos := user_repo.NewRepositories(pool, nil, nil)
	cardRepos := card_repo.NewRepositories(pool, nil)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	saldoService := service.NewService(&service.Deps{
		Repositories: saldoRepos,
		CardAdapter:  s.ts.CardAdapter,
		Logger:       log,
		Cache:        cacheStore,
	})

	// Seed user + card so saldo creation has a valid card to attach to.
	user, err := userRepos.UserCommand().CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Graphql",
		LastName:  "Saldo",
		Email:     "graphql.saldo@example.com",
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

	saldoHandler := gapi.NewHandler(saldoService)
	server := grpc.NewServer()
	pb.RegisterSaldoCommandServiceServer(server, saldoHandler)
	pb.RegisterSaldoQueryServiceServer(server, saldoHandler)
	s.grpcServer = server

	lis, err := net.Listen("tcp", ":0")
	s.Require().NoError(err)
	go func() { _ = server.Serve(lis) }()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.gql = tests.NewGraphQLClient(s.T(), &testhelper.ServiceConnections{SaldoClient: conn}, log, s.redisClient)
}

func (s *SaldoGraphqlTestSuite) TearDownSuite() {
	s.conn.Close()
	s.grpcServer.Stop()
	s.redisClient.Close()
	s.dbPool.Close()
	s.ts.Teardown()
}

func (s *SaldoGraphqlTestSuite) TestSaldoGraphqlLifecycle() {
	// 1. Create
	create := s.gql.Do(`
		mutation Create($input: CreateSaldoInput!) {
			createSaldo(input: $input) { status message data { saldo_id card_number total_balance } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"card_number":   s.cardNumber,
		"total_balance": 100000,
	}})
	created := tests.Payload(s.T(), create, "createSaldo")
	s.saldoID = tests.Int(s.T(), created, "saldo_id")
	s.Require().NotZero(s.saldoID)
	s.Equal(s.cardNumber, tests.Str(s.T(), created, "card_number"))

	// 2. FindById
	found := s.gql.Do(`
		query Find($input: FindByIdSaldoInput!) { findByIdSaldo(input: $input) { status data { saldo_id } } }`,
		map[string]interface{}{"input": map[string]interface{}{"saldo_id": s.saldoID}})
	s.Equal(s.saldoID, tests.Int(s.T(), tests.Payload(s.T(), found, "findByIdSaldo"), "saldo_id"))

	// 3. FindByCardNumber
	byCard := s.gql.Do(`
		query ByCard($card: String!) { findByCardNumberSaldo(card_number: $card) { status data { card_number } } }`,
		map[string]interface{}{"card": s.cardNumber})
	s.Equal(s.cardNumber, tests.Str(s.T(), tests.Payload(s.T(), byCard, "findByCardNumberSaldo"), "card_number"))

	// 4. FindAll
	all := s.gql.Do(`query { findAllSaldo(input: {page: 1, page_size: 10}) { status data { saldo_id } } }`, nil)
	s.NotEmpty(tests.Items(s.T(), all, "findAllSaldo"))

	// 5. Update
	update := s.gql.Do(`
		mutation Update($input: UpdateSaldoInput!) {
			updateSaldo(input: $input) { status data { saldo_id total_balance } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"saldo_id":      s.saldoID,
		"card_number":   s.cardNumber,
		"total_balance": 250000,
	}})
	s.Equal(250000, tests.Int(s.T(), tests.Payload(s.T(), update, "updateSaldo"), "total_balance"))

	// 6. Trash / restore
	trashed := s.gql.Do(`
		mutation Trash($input: FindByIdSaldoInput!) { trashedSaldo(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"saldo_id": s.saldoID}})
	s.Equal("success", tests.Status(s.T(), trashed, "trashedSaldo"))

	trashedList := s.gql.Do(`query { findByTrashedSaldo(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), trashedList, "findByTrashedSaldo"))

	restored := s.gql.Do(`
		mutation Restore($input: FindByIdSaldoInput!) { restoreSaldo(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"saldo_id": s.saldoID}})
	s.Equal("success", tests.Status(s.T(), restored, "restoreSaldo"))

	// 7. DeletePermanent
	s.gql.Do(`
		mutation Trash($input: FindByIdSaldoInput!) { trashedSaldo(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"saldo_id": s.saldoID}})
	permanent := s.gql.Do(`
		mutation DeletePermanent($input: FindByIdSaldoInput!) { deleteSaldoPermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"saldo_id": s.saldoID}})
	s.Equal("success", tests.Status(s.T(), permanent, "deleteSaldoPermanent"))

	// 8. Bulk
	restoreAll := s.gql.Do(`mutation { restoreAllSaldo { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), restoreAll, "restoreAllSaldo"))
	deleteAll := s.gql.Do(`mutation { deleteAllSaldoPermanent { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), deleteAll, "deleteAllSaldoPermanent"))
}

func (s *SaldoGraphqlTestSuite) TestSaldoGraphqlStats() {
	year := time.Now().Year()

	cases := []struct {
		field string
		query string
		vars  map[string]interface{}
	}{
		{"findMonthlySaldoBalanceStats", `query($i: FindYearStatsInput!){ findMonthlySaldoBalanceStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findYearlySaldoBalanceStats", `query($i: FindYearStatsInput!){ findYearlySaldoBalanceStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findMonthlySaldoTotalStats", `query($i: FindYearStatsInput!){ findMonthlySaldoTotalStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findYearlySaldoTotalStats", `query($i: FindYearStatsInput!){ findYearlySaldoTotalStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
	}
	for _, tc := range cases {
		data := s.gql.Do(tc.query, tc.vars)
		s.Equal("success", tests.Status(s.T(), data, tc.field), tc.field)
	}
}

func TestSaldoGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(SaldoGraphqlTestSuite))
}
