package repository

import (
	"context"

	pbai "github.com/MamangRust/microservice-payment-gateway-grpc/pb/ai_security"
	pbcard "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	pbsaldo "github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/outbox"
	"github.com/jmoiron/sqlx"
)

type Repositories interface {
	TransferQueryRepository
	TransferCommandRepository
	IdempotencyRepository
	OutboxRepository

	// The adapters below resolve cross-domain data through the saldo/card/
	// ai-security services' gRPC APIs under their dependency guards. They are
	// built inside NewRepositories so the transfer service never reads those
	// services' tables directly.
	SaldoAdapter() adapter.SaldoAdapter
	CardAdapter() adapter.CardAdapter
	AISecurityAdapter() adapter.AISecurityAdapter
}

type repositories struct {
	TransferQueryRepository
	TransferCommandRepository
	IdempotencyRepository
	OutboxRepository

	saldo      adapter.SaldoAdapter
	card       adapter.CardAdapter
	aiSecurity adapter.AISecurityAdapter
}

// GuardOptions collects the dependency guards applied to the saldo/card/
// ai-security adapters built inside NewRepositories.
type GuardOptions struct {
	Saldo      []adapter.GuardOption
	Card       []adapter.GuardOption
	AISecurity []adapter.GuardOption
}

// Deps carries the raw gRPC clients and guards NewRepositories uses to build the
// cross-domain adapters.
type Deps struct {
	DB *sqlx.DB

	SaldoQueryClient   pbsaldo.SaldoQueryServiceClient
	SaldoCommandClient pbsaldo.SaldoCommandServiceClient
	CardQueryClient    pbcard.CardQueryServiceClient
	CardCommandClient  pbcard.CardCommandServiceClient
	AISecurityClient   pbai.AISecurityServiceClient

	Guards GuardOptions
}

// NewRepositories assembles the transfer-owned SQL repositories and builds the
// shared saldo/card/ai-security gRPC adapters from the raw clients under their
// guards.
func NewRepositories(deps *Deps) Repositories {
	g := deps.Guards

	return &repositories{
		TransferQueryRepository:   NewTransferQueryRepository(deps.DB),
		TransferCommandRepository: NewTransferCommandRepository(deps.DB),
		IdempotencyRepository:     NewTransferIdempotencyRepository(deps.DB),
		OutboxRepository:          outbox.NewStore(storeInsert(deps.DB)),
		saldo:                     adapter.NewSaldoAdapter(deps.SaldoQueryClient, deps.SaldoCommandClient, g.Saldo...),
		card:                      adapter.NewCardAdapter(deps.CardQueryClient, deps.CardCommandClient, g.Card...),
		aiSecurity:                adapter.NewAISecurityAdapter(deps.AISecurityClient, g.AISecurity...),
	}
}

func (r *repositories) SaldoAdapter() adapter.SaldoAdapter {
	return r.saldo
}

func (r *repositories) CardAdapter() adapter.CardAdapter {
	return r.card
}

func (r *repositories) AISecurityAdapter() adapter.AISecurityAdapter {
	return r.aiSecurity
}

// storeInsert adapts the package-level InsertOutboxDB to the shared
// outbox.Store insert-func contract.
func storeInsert(database *sqlx.DB) func(context.Context, db.OutboxRecord) error {
	return func(ctx context.Context, record db.OutboxRecord) error {
		return db.InsertOutboxDB(ctx, database, record)
	}
}
