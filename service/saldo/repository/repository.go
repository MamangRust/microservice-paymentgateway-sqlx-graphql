package repository

import (
	pbcard "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/jmoiron/sqlx"
)

type Repositories interface {
	SaldoQueryRepository
	SaldoCommandRepository

	// CardAdapter resolves card data through the card service gRPC API under its
	// dependency guard. It is built inside NewRepositories so the saldo service
	// never reads the cards table directly.
	CardAdapter() adapter.CardAdapter
}

type repositories struct {
	SaldoQueryRepository
	SaldoCommandRepository

	card adapter.CardAdapter
}

// GuardOptions collects the dependency guards applied to the card adapter built
// inside NewRepositories.
type GuardOptions struct {
	Card []adapter.GuardOption
}

// NewRepositories assembles the saldo-owned SQL repositories and builds the
// shared card gRPC adapter from the raw clients under their guards.
func NewRepositories(
	database *sqlx.DB,
	cardQueryClient pbcard.CardQueryServiceClient,
	cardCommandClient pbcard.CardCommandServiceClient,
	guards ...GuardOptions,
) Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	return &repositories{
		SaldoQueryRepository:   NewSaldoQueryRepository(database),
		SaldoCommandRepository: NewSaldoCommandRepository(database),
		card:                   adapter.NewCardAdapter(cardQueryClient, cardCommandClient, g.Card...),
	}
}

func (r *repositories) CardAdapter() adapter.CardAdapter {
	return r.card
}
