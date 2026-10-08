package repository

import (
	pbuser "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/jmoiron/sqlx"
)

// Repositories contains all repositories used by the card service.
type Repositories struct {
	CardCommand         CardCommandRepository
	CardQuery           CardQueryRepository
	CardAuthTransaction CardAuthTransactionRepository
	CardPayment         CardPaymentRepository
	CardReward          CardRewardRepository
	BillingCycle        BillingCycleRepository

	// UserAdapter resolves card owners through the user service gRPC API under
	// its dependency guard. It is built here so the card service never reaches
	// into the users table directly.
	UserAdapter adapter.UserAdapter
}

// GuardOptions collects the dependency guards applied to the user adapter built
// inside NewRepositories.
type GuardOptions struct {
	User []adapter.GuardOption
}

// NewRepositories assembles the card-owned SQL repositories and builds the
// shared user gRPC adapter from the raw client under its guard.
func NewRepositories(database *sqlx.DB, userQueryClient pbuser.UserQueryServiceClient, guards ...GuardOptions) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	return &Repositories{
		CardQuery:           NewCardQueryRepository(database),
		CardCommand:         NewCardCommandRepository(database),
		CardAuthTransaction: NewCardAuthTransactionRepository(database),
		CardPayment:         NewCardPaymentRepository(database),
		CardReward:          NewCardRewardRepository(database),
		BillingCycle:        NewBillingCycleRepository(database),
		UserAdapter:         adapter.NewUserAdapter(userQueryClient, g.User...),
	}
}
