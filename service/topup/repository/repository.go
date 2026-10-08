package repository

import (
	"context"

	pbcard "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	pbsaldo "github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/topup/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/outbox"
	"github.com/jmoiron/sqlx"
)

type Repositories interface {
	TopupQueryRepository
	TopupCommandRepository
	IdempotencyRepository
	OutboxRepository

	// CardAdapter and SaldoAdapter resolve cross-domain data through the
	// card/saldo services' gRPC APIs under their dependency guards. They are
	// built inside NewRepositories so the topup service never reads the
	// cards/saldos tables directly.
	CardAdapter() adapter.CardAdapter
	SaldoAdapter() adapter.SaldoAdapter
}

type repositories struct {
	TopupQueryRepository
	TopupCommandRepository
	IdempotencyRepository
	OutboxRepository

	card  adapter.CardAdapter
	saldo adapter.SaldoAdapter
}

// GuardOptions collects the dependency guards applied to the card/saldo adapters
// built inside NewRepositories.
type GuardOptions struct {
	Card  []adapter.GuardOption
	Saldo []adapter.GuardOption
}

// NewRepositories assembles the topup-owned SQL repositories and builds the
// shared card/saldo gRPC adapters from the raw clients under their guards.
func NewRepositories(
	database *sqlx.DB,
	cardQueryClient pbcard.CardQueryServiceClient,
	cardCommandClient pbcard.CardCommandServiceClient,
	saldoQueryClient pbsaldo.SaldoQueryServiceClient,
	saldoCommandClient pbsaldo.SaldoCommandServiceClient,
	guards ...GuardOptions,
) Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	return &repositories{
		TopupQueryRepository:   NewTopupQueryRepository(database),
		TopupCommandRepository: NewTopupCommandRepository(database),
		IdempotencyRepository:  NewTopupIdempotencyRepository(database),
		OutboxRepository: outbox.NewStore(func(ctx context.Context, record db.OutboxRecord) error {
			return db.InsertOutboxDB(ctx, database, record)
		}),
		card:  adapter.NewCardAdapter(cardQueryClient, cardCommandClient, g.Card...),
		saldo: adapter.NewSaldoAdapter(saldoQueryClient, saldoCommandClient, g.Saldo...),
	}
}

func (r *repositories) CardAdapter() adapter.CardAdapter {
	return r.card
}

func (r *repositories) SaldoAdapter() adapter.SaldoAdapter {
	return r.saldo
}
