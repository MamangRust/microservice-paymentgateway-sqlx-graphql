package transfer_test

import (
	"context"
	"net"
	"testing"
	"time"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transfer"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/testhelper"
	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	saldo_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer/service"
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

type TransferGraphqlTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	dbPool      *sqlx.DB
	redisClient *redis.Client
	grpcServer  *grpc.Server
	conn        *grpc.ClientConn
	gql         *tests.GraphQLClient
	senderCard  string
	receiverCard string
	transferID  int
}

func (s *TransferGraphqlTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card", "saldo", "transfer"))

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

	transferService := service.NewService(&service.Deps{
		Kafka:            nil,
		Repositories:     repository.NewRepositories(&repository.Deps{DB: pool}),
		CardAdapter:      s.ts.CardAdapter,
		SaldoAdapter:     s.ts.SaldoAdapter,
		Logger:           log,
		Cache:            cacheStore,
		AISecurityClient: nil,
	})

	// Sender with a funded card.
	sender, err := userRepos.UserCommand().CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Graphql",
		LastName:  "Sender",
		Email:     "graphql.sender@example.com",
		Password:  "password123",
	})
	s.Require().NoError(err)
	sCard, err := cardRepos.CardCommand.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID:       int(sender.UserID),
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(1, 0, 0),
		CVV:          "111",
		CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.senderCard = sCard.CardNumber
	_, err = saldoRepos.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{
		CardNumber:   s.senderCard,
		TotalBalance: 1000000,
	})
	s.Require().NoError(err)

	// Receiver with an empty card.
	receiver, err := userRepos.UserCommand().CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Graphql",
		LastName:  "Receiver",
		Email:     "graphql.receiver@example.com",
		Password:  "password123",
	})
	s.Require().NoError(err)
	rCard, err := cardRepos.CardCommand.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID:       int(receiver.UserID),
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(1, 0, 0),
		CVV:          "222",
		CardProvider: "mastercard",
	})
	s.Require().NoError(err)
	s.receiverCard = rCard.CardNumber
	_, err = saldoRepos.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{
		CardNumber:   s.receiverCard,
		TotalBalance: 0,
	})
	s.Require().NoError(err)

	transferHandler := handler.NewHandler(transferService)
	server := grpc.NewServer()
	pb.RegisterTransferCommandServiceServer(server, transferHandler)
	pb.RegisterTransferQueryServiceServer(server, transferHandler)
	s.grpcServer = server

	lis, err := net.Listen("tcp", ":0")
	s.Require().NoError(err)
	go func() { _ = server.Serve(lis) }()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.gql = tests.NewGraphQLClient(s.T(), &testhelper.ServiceConnections{TransferClient: conn}, log, s.redisClient)
}

func (s *TransferGraphqlTestSuite) TearDownSuite() {
	s.conn.Close()
	s.grpcServer.Stop()
	s.redisClient.Close()
	s.dbPool.Close()
	s.ts.Teardown()
}

func (s *TransferGraphqlTestSuite) TestTransferGraphqlLifecycle() {
	// 1. Create
	create := s.gql.Do(`
		mutation Create($input: CreateTransferInput!) {
			createTransfer(input: $input) { status message data { id transfer_from transfer_to transfer_amount } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"transfer_from":   s.senderCard,
		"transfer_to":     s.receiverCard,
		"transfer_amount": 100000,
	}})
	created := tests.Payload(s.T(), create, "createTransfer")
	s.transferID = tests.Int(s.T(), created, "id")
	s.Require().NotZero(s.transferID)
	s.Equal(s.senderCard, tests.Str(s.T(), created, "transfer_from"))
	s.Equal(s.receiverCard, tests.Str(s.T(), created, "transfer_to"))

	// 2. FindById
	found := s.gql.Do(`
		query Find($input: FindByIdTransferInput!) { findByIdTransfer(input: $input) { status data { id } } }`,
		map[string]interface{}{"input": map[string]interface{}{"transfer_id": s.transferID}})
	s.Equal(s.transferID, tests.Int(s.T(), tests.Payload(s.T(), found, "findByIdTransfer"), "id"))

	// 3. FindByTransferFrom / FindByTransferTo
	byFrom := s.gql.Do(`
		query ByFrom($input: FindTransferByTransferFromInput!) { findTransferByTransferFrom(input: $input) { status data { transfer_from } } }`,
		map[string]interface{}{"input": map[string]interface{}{"transfer_from": s.senderCard}})
	s.NotEmpty(tests.Items(s.T(), byFrom, "findTransferByTransferFrom"))

	byTo := s.gql.Do(`
		query ByTo($input: FindTransferByTransferToInput!) { findTransferByTransferTo(input: $input) { status data { transfer_to } } }`,
		map[string]interface{}{"input": map[string]interface{}{"transfer_to": s.receiverCard}})
	s.NotEmpty(tests.Items(s.T(), byTo, "findTransferByTransferTo"))

	// 4. FindAll
	all := s.gql.Do(`query { findAllTransfer(input: {page: 1, page_size: 10}) { status data { id } } }`, nil)
	s.NotEmpty(tests.Items(s.T(), all, "findAllTransfer"))

	// 5. Update
	update := s.gql.Do(`
		mutation Update($input: UpdateTransferInput!) {
			updateTransfer(input: $input) { status data { id transfer_amount } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"transfer_id":     s.transferID,
		"transfer_from":   s.senderCard,
		"transfer_to":     s.receiverCard,
		"transfer_amount": 150000,
	}})
	s.Equal(150000, tests.Int(s.T(), tests.Payload(s.T(), update, "updateTransfer"), "transfer_amount"))

	// 6. Trash / restore
	trashed := s.gql.Do(`
		mutation Trash($input: FindByIdTransferInput!) { trashedTransfer(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"transfer_id": s.transferID}})
	s.Equal("success", tests.Status(s.T(), trashed, "trashedTransfer"))

	trashedList := s.gql.Do(`query { findByTrashedTransfer(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), trashedList, "findByTrashedTransfer"))

	restored := s.gql.Do(`
		mutation Restore($input: FindByIdTransferInput!) { restoreTransfer(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"transfer_id": s.transferID}})
	s.Equal("success", tests.Status(s.T(), restored, "restoreTransfer"))

	// 7. DeletePermanent
	s.gql.Do(`
		mutation Trash($input: FindByIdTransferInput!) { trashedTransfer(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"transfer_id": s.transferID}})
	permanent := s.gql.Do(`
		mutation DeletePermanent($input: FindByIdTransferInput!) { deleteTransferPermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"transfer_id": s.transferID}})
	s.Equal("success", tests.Status(s.T(), permanent, "deleteTransferPermanent"))

	// 8. Bulk
	restoreAll := s.gql.Do(`mutation { restoreAllTransfer { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), restoreAll, "restoreAllTransfer"))
	deleteAll := s.gql.Do(`mutation { deleteAllTransferPermanent { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), deleteAll, "deleteAllTransferPermanent"))
}

// TestTransferGraphqlStats covers the transfer-scoped stats queries. The gateway
// resolves these from the stats-reader service; in the harness they are wired to
// the stub surface, so the assertion is on the response envelope.
func (s *TransferGraphqlTestSuite) TestTransferGraphqlStats() {
	year := time.Now().Year()
	month := int(time.Now().Month())

	cases := []struct {
		field string
		query string
		vars  map[string]interface{}
	}{
		{"findMonthlyTransferAmountStats", `query($i: FindYearStatsInput!){ findMonthlyTransferAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findYearlyTransferAmountStats", `query($i: FindYearStatsInput!){ findYearlyTransferAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findMonthlyTransferSenderAmountStats", `query($i: FindYearCardNumberStatsInput!){ findMonthlyTransferSenderAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "card_number": s.senderCard}}},
		{"findYearlyTransferSenderAmountStats", `query($i: FindYearCardNumberStatsInput!){ findYearlyTransferSenderAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "card_number": s.senderCard}}},
		{"findMonthlyTransferReceiverAmountStats", `query($i: FindYearCardNumberStatsInput!){ findMonthlyTransferReceiverAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "card_number": s.receiverCard}}},
		{"findYearlyTransferReceiverAmountStats", `query($i: FindYearCardNumberStatsInput!){ findYearlyTransferReceiverAmountStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "card_number": s.receiverCard}}},

		{"findMonthlyTransferStatusSuccessStats", `query($i: FindMonthlyYearStatsInput!){ findMonthlyTransferStatusSuccessStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "month": month}}},
		{"findYearlyTransferStatusSuccessStats", `query($i: FindYearStatsInput!){ findYearlyTransferStatusSuccessStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
		{"findMonthlyTransferStatusFailedStats", `query($i: FindMonthlyYearStatsInput!){ findMonthlyTransferStatusFailedStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year, "month": month}}},
		{"findYearlyTransferStatusFailedStats", `query($i: FindYearStatsInput!){ findYearlyTransferStatusFailedStats(input:$i){ status } }`, map[string]interface{}{"i": map[string]interface{}{"year": year}}},
	}
	for _, tc := range cases {
		data := s.gql.Do(tc.query, tc.vars)
		s.Equal("success", tests.Status(s.T(), data, tc.field), tc.field)
	}
}

func TestTransferGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransferGraphqlTestSuite))
}
