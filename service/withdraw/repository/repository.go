package repository

import (
	"context"

	pbai "github.com/MamangRust/microservice-payment-gateway-grpc/pb/ai_security"
	pbcard "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	pbsaldo "github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/outbox"
	"github.com/jmoiron/sqlx"
)

type Repositories interface {
	WithdrawQueryRepository
	WithdrawCommandRepository
	IdempotencyRepository
	OutboxRepository

	// The adapters below resolve cross-domain data through the card/saldo/
	// ai-security services' gRPC APIs under their dependency guards. They are
	// built inside NewRepositories so the withdraw service never reads those
	// services' tables directly.
	CardAdapter() adapter.CardAdapter
	SaldoAdapter() adapter.SaldoAdapter
	AISecurityAdapter() adapter.AISecurityAdapter
}

type repositories struct {
	WithdrawQueryRepository
	WithdrawCommandRepository
	IdempotencyRepository
	OutboxRepository

	card       adapter.CardAdapter
	saldo      adapter.SaldoAdapter
	aiSecurity adapter.AISecurityAdapter
}

// GuardOptions collects the dependency guards applied to the card/saldo/
// ai-security adapters built inside NewRepositories.
type GuardOptions struct {
	Card       []adapter.GuardOption
	Saldo      []adapter.GuardOption
	AISecurity []adapter.GuardOption
}

// Deps carries the raw gRPC clients and guards NewRepositories uses to build the
// cross-domain adapters.
type Deps struct {
	DB *sqlx.DB

	CardQueryClient    pbcard.CardQueryServiceClient
	CardCommandClient  pbcard.CardCommandServiceClient
	SaldoQueryClient   pbsaldo.SaldoQueryServiceClient
	SaldoCommandClient pbsaldo.SaldoCommandServiceClient
	AISecurityClient   pbai.AISecurityServiceClient

	Guards GuardOptions
}

// NewRepositories assembles the withdraw-owned SQL repositories and builds the
// shared card/saldo/ai-security gRPC adapters from the raw clients under their
// guards.
func NewRepositories(deps *Deps) Repositories {
	g := deps.Guards

	return &repositories{
		WithdrawQueryRepository:   NewWithdrawQueryRepository(deps.DB),
		WithdrawCommandRepository: NewWithdrawCommandRepository(deps.DB),
		IdempotencyRepository:     NewWithdrawIdempotencyRepository(deps.DB),
		OutboxRepository: outbox.NewStore(func(ctx context.Context, record db.OutboxRecord) error {
			return db.InsertOutboxDB(ctx, deps.DB, record)
		}),
		card:       adapter.NewCardAdapter(deps.CardQueryClient, deps.CardCommandClient, g.Card...),
		saldo:      adapter.NewSaldoAdapter(deps.SaldoQueryClient, deps.SaldoCommandClient, g.Saldo...),
		aiSecurity: adapter.NewAISecurityAdapter(deps.AISecurityClient, g.AISecurity...),
	}
}

func (r *repositories) CardAdapter() adapter.CardAdapter {
	return r.card
}

func (r *repositories) SaldoAdapter() adapter.SaldoAdapter {
	return r.saldo
}

func (r *repositories) AISecurityAdapter() adapter.AISecurityAdapter {
	return r.aiSecurity
}
