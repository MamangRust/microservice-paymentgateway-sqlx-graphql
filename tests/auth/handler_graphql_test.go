package auth_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/auth"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/hash"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/testhelper"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/auth/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/auth/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/auth/service"
	tests "github.com/MamangRust/microservice-payment-gateway-test"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthGraphqlTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	dbPool      *sqlx.DB
	redisClient *redis.Client
	grpcServer  *grpc.Server
	conn        *grpc.ClientConn
	gql         *tests.GraphQLClient
	email       string
	password    string
	userID      int
}

func (s *AuthGraphqlTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth"))

	pool, err := sqlx.ConnectContext(s.ts.Ctx, "pgx", s.ts.DBURL)
	s.Require().NoError(err)
	s.dbPool = pool

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	repos := repository.NewRepositories(&repository.RepositoriesDeps{
		DB:                pool,
		UserQueryClient:   s.ts.UserQueryClient,
		UserCommandClient: s.ts.UserCommandClient,
		RoleQueryClient:   s.ts.RoleQueryClient,
		UserRoleClient:    s.ts.UserRoleClient,
	})

	tokenManager, _ := auth.NewManager("mysecret")
	hasher := hash.NewHashingPassword()

	svc := service.NewService(&service.Deps{
		Repositories: repos,
		Logger:       s.ts.Logger,
		Cache:        s.ts.CacheStore,
		Token:        tokenManager,
		Hash:         hasher,
		Kafka:        nil,
	})

	h := handler.NewAuthHandleGrpc(svc, s.ts.Logger)
	server := grpc.NewServer()
	pb.RegisterAuthServiceServer(server, h)
	s.grpcServer = server

	lis, err := net.Listen("tcp", "localhost:0")
	s.Require().NoError(err)
	go func() { _ = server.Serve(lis) }()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn

	// Seed ROLE_ADMIN so registration can grant it.
	_, _ = pool.ExecContext(context.Background(), "INSERT INTO roles (role_name) VALUES ('ROLE_ADMIN')")

	s.email = fmt.Sprintf("graphql.auth.%d@example.com", time.Now().UnixNano())
	s.password = "password123"

	s.gql = tests.NewGraphQLClient(s.T(), &testhelper.ServiceConnections{AuthClient: conn}, s.ts.Logger, s.redisClient)
}

func (s *AuthGraphqlTestSuite) TearDownSuite() {
	s.conn.Close()
	s.grpcServer.Stop()
	s.redisClient.Close()
	s.dbPool.Close()
	s.ts.Teardown()
}

func (s *AuthGraphqlTestSuite) TestAuthGraphqlLifecycle() {
	// 1. Register
	register := s.gql.Do(`
		mutation Register($input: RegisterInput!) {
			registerUser(input: $input) { status message data { id email } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"firstname":        "Graphql",
		"lastname":         "Auth",
		"email":            s.email,
		"password":         s.password,
		"confirm_password": s.password,
	}})
	registered := tests.Payload(s.T(), register, "registerUser")
	s.userID = tests.Int(s.T(), registered, "id")
	s.Require().NotZero(s.userID)
	s.Equal(s.email, tests.Str(s.T(), registered, "email"))

	// 2. Login
	login := s.gql.Do(`
		mutation Login($input: LoginInput!) {
			loginUser(input: $input) { status message data { access_token refresh_token } }
		}`, map[string]interface{}{"input": map[string]interface{}{
		"email":    s.email,
		"password": s.password,
	}})
	tokens := tests.Payload(s.T(), login, "loginUser")
	accessToken := tests.Str(s.T(), tokens, "access_token")
	s.Require().NotEmpty(accessToken)
	refreshToken := tests.Str(s.T(), tokens, "refresh_token")
	s.Require().NotEmpty(refreshToken)

	// 3. GetMe — the resolver reads the caller identity from the request
	// context (set by AuthMiddleware), so inject the registered user.
	me := s.gql.WithUser(s.userID).Do(`
		query Me($input: GetMeInput!) { getMe(input: $input) { status data { id email } } }`,
		map[string]interface{}{"input": map[string]interface{}{"access_token": accessToken}})
	meData := tests.Payload(s.T(), me, "getMe")
	s.Equal(s.email, tests.Str(s.T(), meData, "email"))

	// 4. RefreshToken
	refresh := s.gql.Do(`
		mutation Refresh($input: RefreshTokenInput!) {
			refreshToken(input: $input) { status data { access_token } }
		}`, map[string]interface{}{"input": map[string]interface{}{"refresh_token": refreshToken}})
	s.NotEmpty(tests.Str(s.T(), tests.Payload(s.T(), refresh, "refreshToken"), "access_token"))
}

func TestAuthGraphqlSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(AuthGraphqlTestSuite))
}
