package testhelper

import (
	"context"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	mycontext "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/context"
	graph "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/middlewares"
	mencache "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/internal/redis"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/redis/go-redis/v9"
	"github.com/vektah/gqlparser/v2/ast"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ServiceConnections mirrors graph.ServiceConnections for test use.
type ServiceConnections = graph.ServiceConnections

// CreateDummyConn creates a lazy gRPC connection that will never actually connect.
func CreateDummyConn() *grpc.ClientConn {
	conn, err := grpc.NewClient("localhost:1", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil
	}
	return conn
}

// NewResolver creates a Resolver with the provided service connections.
func NewResolver(conns *ServiceConnections, log logger.LoggerInterface) *graph.Resolver {
	myMencache := mencache.NewCacheApiGateway(&mencache.Deps{
		Redis:  nil,
		Logger: log,
	})

	return graph.NewResolver(&graph.Deps{
		Clients:  conns,
		Logger:   log,
		Kafka:    nil,
		Mencache: myMencache,
	})
}

// NewResolverWithRedis creates a Resolver with the provided service connections and Redis client.
func NewResolverWithRedis(conns *ServiceConnections, log logger.LoggerInterface, redisClient *redis.Client) *graph.Resolver {
	myMencache := mencache.NewCacheApiGateway(&mencache.Deps{
		Redis:  redisClient,
		Logger: log,
	})

	return graph.NewResolver(&graph.Deps{
		Clients:  conns,
		Logger:   log,
		Kafka:    nil,
		Mencache: myMencache,
	})
}

// permissiveRoleChecker stands in for the RBAC checker of the running gateway.
//
// The harness mounts the schema without AuthMiddleware, and that middleware is
// what puts the authenticated user in the request context, so there is no user
// for the @hasRole directive to authorise and it passes every field through.
// The checker is wired anyway so the directive is never left nil, and so that a
// test which injects a user (see WithUser) gets a predictable answer instead of
// an "rbac: role checker is not configured" error.
type permissiveRoleChecker struct{}

func (permissiveRoleChecker) CheckRole(context.Context, int, ...string) error { return nil }

// NewGraphQLHTTPHandler creates an http.Handler from a gqlgen Resolver.
func NewGraphQLHTTPHandler(resolver *graph.Resolver) http.Handler {
	return NewGraphQLHTTPHandlerWithRoleChecker(resolver, permissiveRoleChecker{})
}

// NewGraphQLHTTPHandlerWithRoleChecker builds the handler with an explicit RBAC
// checker, so a test can exercise the @hasRole directive (pair it with WithUser
// to simulate an authenticated caller).
func NewGraphQLHTTPHandlerWithRoleChecker(resolver *graph.Resolver, checker middlewares.RoleChecker) http.Handler {
	srv := handler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers:  resolver,
		Directives: graph.DirectiveRoot{HasRole: middlewares.HasRole(checker)},
	}))

	srv.AddTransport(transport.POST{})
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.Introspection{})

	return srv
}

// WithUser injects the given user id into the request context, emulating what
// AuthMiddleware does for authenticated traffic.
func WithUser(next http.Handler, userID int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(mycontext.WithUserID(r.Context(), userID)))
	})
}

// SeedMerchantCache writes a merchant ID-to-API-key mapping into Redis
// so the permission validation can find it without Kafka.
func SeedMerchantCache(redisClient *redis.Client, merchantID string, apiKey string) error {
	key := "merchant_api_key:" + merchantID
	return redisClient.Set(context.Background(), key, apiKey, 0).Err()
}

// SeedRoleCache writes a user role mapping into Redis
// so the permission validation can find it without Kafka.
func SeedRoleCache(redisClient *redis.Client, userID string, roles []string) error {
	key := "user_roles:" + userID
	return redisClient.Set(context.Background(), key, roles, 0).Err()
}
