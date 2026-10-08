package apps

import (
	"context"
	"fmt"
	"time"

	pbcard "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/kafka"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/resilience"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/server"
	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/handler"
	saldokafka "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/kafka"
	mencache "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/redis"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/service"
	ledgerworker "github.com/MamangRust/microservice-payment-gateway-grpc/shared/ledger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/outbox"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewServer(cfg *server.Config) (*server.GRPCServer, error) {
	srv, err := server.New(cfg)
	if err != nil {
		return nil, err
	}

	queries := srv.DB

	// gRPC Clients for cross-service communication
	connCard, err := grpc.NewClient(viper.GetString("GRPC_CARD_ADDR"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Card service: %w", err)
	}

	cardClientQuery := pbcard.NewCardQueryServiceClient(connCard)
	cardClientCmd := pbcard.NewCardCommandServiceClient(connCard)

	repos := repository.NewRepositories(queries,
		cardClientQuery,
		cardClientCmd,
		repository.GuardOptions{
			Card: []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("card", 5, 30, 100, 3*time.Second, srv.Logger))},
		})

	mykafka, err := kafka.NewKafka(srv.Logger, []string{viper.GetString("KAFKA_BROKERS")})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	svc := service.NewService(&service.Deps{
		Cache:        srv.CacheStore,
		Logger:       srv.Logger,
		Repositories: repos,
		CardAdapter:  repos.CardAdapter(),
		Kafka:        mykafka,
		Outbox:       outbox.NewStore(outboxInsertFunc(srv.DB)),
	})

	kafkaHandler := saldokafka.NewSaldoKafkaHandler(svc, srv.Logger, context.Background())
	err = mykafka.StartConsumers([]string{"saldo-service-topic-create-saldo"}, "saldo-service-group", kafkaHandler)
	if err != nil {
		srv.Logger.Error("Failed to start kafka consumers", zap.Error(err))
	}

	outboxPub := outbox.NewPublisher(outbox.FromDB(srv.DB), mykafka, srv.Logger)
	go outboxPub.Start(srv.Ctx)

	saldoCache := mencache.NewMencache(srv.CacheStore)
	cacheInvalidator := saldokafka.NewSaldoCacheInvalidator(saldoCache, srv.Logger)
	if err := mykafka.StartConsumers([]string{"stats-topic-saldo-events"}, "saldo-cache-invalidator", cacheInvalidator); err != nil {
		srv.Logger.Error("Failed to start saldo cache invalidator consumer", zap.Error(err))
	}

	h := handler.NewHandler(svc)

	srv.RegisterServices = func(gs *grpc.Server) {
		pb.RegisterSaldoQueryServiceServer(gs, h)
		pb.RegisterSaldoCommandServiceServer(gs, h)
	}

	reconciler := ledgerworker.NewReconciler(ledgerReconcilerStore{db: srv.DB}, srv.Logger, 5*time.Minute, 100)
	go reconciler.Start(srv.Ctx)

	return srv, nil
}

// outboxInsertFunc adapts the schema-level InsertOutboxDB to the shared
// outbox.Store insert signature.
func outboxInsertFunc(database *sqlx.DB) func(context.Context, db.OutboxRecord) error {
	return func(ctx context.Context, record db.OutboxRecord) error {
		return db.InsertOutboxDB(ctx, database, record)
	}
}

// ledgerReconcilerStore adapts the schema-level reconciliation queries to the
// shared ledgerworker.ReconciliationStore contract.
type ledgerReconcilerStore struct {
	db *sqlx.DB
}

func (s ledgerReconcilerStore) ListReconciliationMismatches(ctx context.Context, limit int32) ([]*db.ReconciliationRow, error) {
	return db.ListReconciliationMismatches(ctx, s.db, limit)
}

func (s ledgerReconcilerStore) EnqueueReconciliationMismatch(ctx context.Context, item *db.ReconciliationRow) error {
	return db.EnqueueReconciliationMismatch(ctx, s.db, item)
}
