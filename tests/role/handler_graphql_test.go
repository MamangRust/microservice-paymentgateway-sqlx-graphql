package role_test

import (
	"net"
	"testing"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/role"
	pbuserrole "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user_role"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/testhelper"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/role/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/role/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/role/service"
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

type RoleGraphqlTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	dbPool      *sqlx.DB
	redisClient *redis.Client
	grpcServer  *grpc.Server
	conn        *grpc.ClientConn
	gql         *tests.GraphQLClient
	roleID      int
}

func (s *RoleGraphqlTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	s.Require().NoError(s.ts.RunMigrations("user", "role"))

	pool, err := sqlx.ConnectContext(s.ts.Ctx, "pgx", s.ts.DBURL)
	s.Require().NoError(err)
	s.dbPool = pool

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	repos := repository.NewRepositories(pool)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)

	roleService := service.NewService(&service.Deps{
		Repositories: repos,
		Logger:       log,
		Cache:        cacheStore,
	})

	roleHandler := handler.NewHandler(roleService)
	server := grpc.NewServer()
	pb.RegisterRoleCommandServiceServer(server, roleHandler.RoleCommand)
	pb.RegisterRoleQueryServiceServer(server, roleHandler.RoleQuery)
	pbuserrole.RegisterUserRoleServiceServer(server, roleHandler.UserRole)
	s.grpcServer = server

	lis, err := net.Listen("tcp", "localhost:0")
	s.Require().NoError(err)
	go func() { _ = server.Serve(lis) }()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	s.gql = tests.NewGraphQLClient(s.T(), &testhelper.ServiceConnections{RoleClient: conn}, log, s.redisClient)
}

func (s *RoleGraphqlTestSuite) TearDownSuite() {
	s.conn.Close()
	s.grpcServer.Stop()
	s.redisClient.Close()
	s.dbPool.Close()
	s.ts.Teardown()
}

func (s *RoleGraphqlTestSuite) TestRoleGraphqlLifecycle() {
	// 1. Create
	create := s.gql.Do(`
		mutation Create($input: CreateRoleInput!) {
			createRole(input: $input) { status message data { id name } }
		}`, map[string]interface{}{"input": map[string]interface{}{"name": "Graphql Role"}})
	created := tests.Payload(s.T(), create, "createRole")
	s.roleID = tests.Int(s.T(), created, "id")
	s.Require().NotZero(s.roleID)
	s.Equal("Graphql Role", tests.Str(s.T(), created, "name"))

	// 2. FindById
	found := s.gql.Do(`
		query Find($input: FindByIdRoleInput!) { findByIdRole(input: $input) { status data { id name } } }`,
		map[string]interface{}{"input": map[string]interface{}{"role_id": s.roleID}})
	s.Equal(s.roleID, tests.Int(s.T(), tests.Payload(s.T(), found, "findByIdRole"), "id"))

	// 3. FindAll
	all := s.gql.Do(`query { findAllRole(input: {page: 1, page_size: 10}) { status data { id } } }`, nil)
	s.NotEmpty(tests.Items(s.T(), all, "findAllRole"))

	// 4. Update
	update := s.gql.Do(`
		mutation Update($input: UpdateRoleInput!) {
			updateRole(input: $input) { status data { id name } }
		}`, map[string]interface{}{"input": map[string]interface{}{"id": s.roleID, "name": "Graphql Role Updated"}})
	s.Equal("Graphql Role Updated", tests.Str(s.T(), tests.Payload(s.T(), update, "updateRole"), "name"))

	// 5. Trash / restore
	trashed := s.gql.Do(`
		mutation Trash($input: FindByIdRoleInput!) { trashedRole(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"role_id": s.roleID}})
	s.Equal("success", tests.Status(s.T(), trashed, "trashedRole"))

	trashedList := s.gql.Do(`query { findByTrashedRole(input: {page: 1, page_size: 10}) { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), trashedList, "findByTrashedRole"))

	restored := s.gql.Do(`
		mutation Restore($input: FindByIdRoleInput!) { restoreRole(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"role_id": s.roleID}})
	s.Equal("success", tests.Status(s.T(), restored, "restoreRole"))

	// 6. DeletePermanent
	s.gql.Do(`
		mutation Trash($input: FindByIdRoleInput!) { trashedRole(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"role_id": s.roleID}})
	permanent := s.gql.Do(`
		mutation DeletePermanent($input: FindByIdRoleInput!) { deleteRolePermanent(input: $input) { status } }`,
		map[string]interface{}{"input": map[string]interface{}{"role_id": s.roleID}})
	s.Equal("success", tests.Status(s.T(), permanent, "deleteRolePermanent"))

	// 7. Bulk
	restoreAll := s.gql.Do(`mutation { restoreAllRole { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), restoreAll, "restoreAllRole"))
	deleteAll := s.gql.Do(`mutation { deleteAllRolePermanent { status } }`, nil)
	s.Equal("success", tests.Status(s.T(), deleteAll, "deleteAllRolePermanent"))
}

func TestRoleGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(RoleGraphqlTestSuite))
}
