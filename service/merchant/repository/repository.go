package repository

import (
	pbtransaction "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transaction"
	pbuser "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/jmoiron/sqlx"
)

type Repositories interface {
	MerchantQueryRepository
	MerchantCommandRepository
	MerchantDocumentQueryRepository
	MerchantDocumentCommandRepository

	// UserAdapter and TransactionAdapter resolve cross-domain data through the
	// user/transaction services' gRPC APIs under their dependency guards. They
	// are built inside NewRepositories so the merchant service never reads the
	// users or transactions tables directly.
	UserAdapter() adapter.UserAdapter
	TransactionAdapter() adapter.TransactionAdapter
}

type repositories struct {
	MerchantQueryRepository
	MerchantCommandRepository
	MerchantDocumentQueryRepository
	MerchantDocumentCommandRepository

	userAdapter        adapter.UserAdapter
	transactionAdapter adapter.TransactionAdapter
}

// GuardOptions collects the dependency guards applied to the user/transaction
// adapters built inside NewRepositories.
type GuardOptions struct {
	User        []adapter.GuardOption
	Transaction []adapter.GuardOption
}

// NewRepositories assembles the merchant-owned SQL repositories and builds the
// shared user/transaction gRPC adapters from the raw clients under their guards.
func NewRepositories(
	db *sqlx.DB,
	userQueryClient pbuser.UserQueryServiceClient,
	transactionQueryClient pbtransaction.TransactionQueryServiceClient,
	guards ...GuardOptions,
) Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	return &repositories{
		MerchantQueryRepository:           NewMerchantQueryRepository(db),
		MerchantCommandRepository:         NewMerchantCommandRepository(db),
		MerchantDocumentQueryRepository:   NewMerchantDocumentQueryRepository(db),
		MerchantDocumentCommandRepository: NewMerchantDocumentCommandRepository(db),
		userAdapter:                       adapter.NewUserAdapter(userQueryClient, g.User...),
		transactionAdapter:                adapter.NewTransactionAdapter(transactionQueryClient, g.Transaction...),
	}
}

func (r *repositories) UserAdapter() adapter.UserAdapter {
	return r.userAdapter
}

func (r *repositories) TransactionAdapter() adapter.TransactionAdapter {
	return r.transactionAdapter
}
