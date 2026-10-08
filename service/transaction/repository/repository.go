package repository

import (
	"context"

	pbai "github.com/MamangRust/microservice-payment-gateway-grpc/pb/ai_security"
	pbcard "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	pbmerchant "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant"
	pbsaldo "github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/transaction/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/outbox"
	"github.com/jmoiron/sqlx"
)

type Repositories interface {
	TransactionQueryRepository
	TransactionCommandRepository
	IdempotencyRepository
	OutboxRepository

	// The adapters below resolve cross-domain data through the saldo/card/
	// merchant/ai-security services' gRPC APIs under their dependency guards.
	// They are built inside NewRepositories so the transaction service never
	// reads those services' tables directly.
	SaldoAdapter() adapter.SaldoAdapter
	CardAdapter() adapter.CardAdapter
	MerchantAdapter() adapter.MerchantAdapter
	AISecurityAdapter() adapter.AISecurityAdapter
}

// repositories is a struct that contains all the repositories used in the transaction service.
type repositories struct {
	TransactionQueryRepository
	TransactionCommandRepository
	IdempotencyRepository
	OutboxRepository

	saldo      adapter.SaldoAdapter
	card       adapter.CardAdapter
	merchant   adapter.MerchantAdapter
	aiSecurity adapter.AISecurityAdapter
}

// GuardOptions collects the dependency guards applied to the saldo/card/
// merchant/ai-security adapters built inside NewRepositories. Each dependency
// gets its own breaker so one failing dependency cannot starve the others.
type GuardOptions struct {
	Saldo      []adapter.GuardOption
	Card       []adapter.GuardOption
	Merchant   []adapter.GuardOption
	AISecurity []adapter.GuardOption
}

// Deps carries the raw gRPC clients and guards NewRepositories uses to build the
// cross-domain adapters.
type Deps struct {
	DB *sqlx.DB

	SaldoQueryClient    pbsaldo.SaldoQueryServiceClient
	SaldoCommandClient  pbsaldo.SaldoCommandServiceClient
	CardQueryClient     pbcard.CardQueryServiceClient
	CardCommandClient   pbcard.CardCommandServiceClient
	MerchantQueryClient pbmerchant.MerchantQueryServiceClient
	AISecurityClient    pbai.AISecurityServiceClient

	Guards GuardOptions
}

// NewRepositories creates a new instance of Repositories with the provided
// database client and builds the saldo/card/merchant/ai-security gRPC adapters
// from the raw clients under their guards.
func NewRepositories(deps *Deps) Repositories {
	g := deps.Guards

	return &repositories{
		TransactionQueryRepository:   NewTransactionQueryRepository(deps.DB),
		TransactionCommandRepository: NewTransactionCommandRepository(deps.DB),
		IdempotencyRepository:        NewTransactionIdempotencyRepository(deps.DB),
		OutboxRepository: outbox.NewStore(func(ctx context.Context, record db.OutboxRecord) error {
			return db.InsertOutboxDB(ctx, deps.DB, record)
		}),
		saldo:      adapter.NewSaldoAdapter(deps.SaldoQueryClient, deps.SaldoCommandClient, g.Saldo...),
		card:       adapter.NewCardAdapter(deps.CardQueryClient, deps.CardCommandClient, g.Card...),
		merchant:   adapter.NewMerchantAdapter(deps.MerchantQueryClient, g.Merchant...),
		aiSecurity: adapter.NewAISecurityAdapter(deps.AISecurityClient, g.AISecurity...),
	}
}

func (r *repositories) SaldoAdapter() adapter.SaldoAdapter {
	return r.saldo
}

func (r *repositories) CardAdapter() adapter.CardAdapter {
	return r.card
}

func (r *repositories) MerchantAdapter() adapter.MerchantAdapter {
	return r.merchant
}

func (r *repositories) AISecurityAdapter() adapter.AISecurityAdapter {
	return r.aiSecurity
}
