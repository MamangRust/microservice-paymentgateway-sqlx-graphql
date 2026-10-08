package user_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/hash"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	gapi "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/user/service"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/testhelper"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"
	tests "github.com/MamangRust/microservice-payment-gateway-test"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserGraphqlTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	dbPool      *sqlx.DB
	redisClient *redis.Client
	grpcServer  *grpc.Server
	conn        *grpc.ClientConn
	gql         *tests.GraphQLClient
	userID      int
	userEmail   string
}

func (s *UserGraphqlTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	s.Require().NoError(s.ts.RunMigrations("user", "role"))
	s.Require().NoError(s.ts.SeedRole("ROLE_ADMIN"))

	pool, err := sqlx.ConnectContext(s.ts.Ctx, "pgx", s.ts.DBURL)
	s.Require().NoError(err)
	s.dbPool = pool

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	repos := repository.NewRepositories(pool, s.ts.RoleClient, s.ts.UserRoleClient)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	hasher := hash.NewHashingPassword()
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	userService := service.NewService(&service.Deps{
		Cache:        cacheStore,
		Repositories: repos,
		Hash:         hasher,
		Logger:       log,
	})

	userHandler := gapi.NewHandler(userService)
	server := grpc.NewServer()
	pb.RegisterUserQueryServiceServer(server, userHandler)
	pb.RegisterUserCommandServiceServer(server, userHandler)
	s.grpcServer = server

	lis, err := net.Listen("tcp", ":0")
	s.Require().NoError(err)
	go func() { _ = server.Serve(lis) }()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.gql = tests.NewGraphQLClient(s.T(), &testhelper.ServiceConnections{UserClient: conn}, log, s.redisClient)
}

func (s *UserGraphqlTestSuite) TearDownSuite() {
	s.conn.Close()
	s.grpcServer.Stop()
	s.redisClient.Close()
	s.dbPool.Close()
	s.ts.Teardown()
}

func (s *UserGraphqlTestSuite) TestUserGraphqlLifecycle() {
	s.userEmail = fmt.Sprintf("graphql.user.%d@example.com", time.Now().UnixNano())

	// 1. Create
	create := s.gql.Do(`
		mutation Create($input: CreateUserInput!) {
			createUser(input: $input) { status message data { id firstname lastname email } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"firstname":        "Graphql",
		"lastname":         "User",
		"email":            s.userEmail,
		"password":         "password123",
		"confirm_password": "password123",
	}})
	created := tests.Payload(s.T(), create, "createUser")
	s.userID = tests.Int(s.T(), created, "id")
	s.Require().NotZero(s.userID)
	s.Equal(s.userEmail, tests.Str(s.T(), created, "email"))

	// 2. FindById
	found := s.gql.Do(`
		query Find($input: FindByIdUserInput!) { findByIdUser(input: $input) { status data { id email } } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": s.userID}})
	s.Equal(s.userID, tests.Int(s.T(), tests.Payload(s.T(), found, "findByIdUser"), "id"))

	// 3. FindAll
	all := s.gql.Do(`
		query { findAllUser(input: {page: 1, page_size: 10}) { status data { id } } }`, nil)
	s.NotEmpty(tests.Items(s.T(), all, "findAllUser"))

	// 4. Update
	update := s.gql.Do(`
		mutation Update($input: UpdateUserInput!) {
			updateUser(input: $input) { status data { id firstname } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"id":               s.userID,
		"firstname":        "Updated",
		"lastname":         "User",
		"email":            s.userEmail,
		"password":         "password123",
		"confirm_password": "password123",
	}})
	s.Equal("Updated", tests.Str(s.T(), tests.Payload(s.T(), update, "updateUser"), "firstname"))

	// 5. Trash
	trashed := s.gql.Do(`
		mutation Trash($input: FindByIdUserInput!) { trashedUser(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": s.userID}})
	s.Equal("success", tests.Status(s.T(), trashed, "trashedUser"))

	// 6. FindByTrashed
	trashedList := s.gql.Do(`
		query { findByTrashedUser(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), trashedList, "findByTrashedUser"))

	// 7. Restore
	restored := s.gql.Do(`
		mutation Restore($input: FindByIdUserInput!) { restoreUser(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": s.userID}})
	s.Equal("success", tests.Status(s.T(), restored, "restoreUser"))

	// 8. DeletePermanent
	s.gql.Do(`
		mutation Trash($input: FindByIdUserInput!) { trashedUser(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": s.userID}})
	permanent := s.gql.Do(`
		mutation DeletePermanent($input: FindByIdUserInput!) { deleteUserPermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"id": s.userID}})
	s.Equal("success", tests.Status(s.T(), permanent, "deleteUserPermanent"))

	// 9. Bulk operations
	restoreAll := s.gql.Do(`mutation { restoreAllUser { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), restoreAll, "restoreAllUser"))
	deleteAll := s.gql.Do(`mutation { deleteAllUserPermanent { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), deleteAll, "deleteAllUserPermanent"))
}

func TestUserGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(UserGraphqlTestSuite))
}
