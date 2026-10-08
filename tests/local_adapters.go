package tests

// Local implementations of the pkg/adapter interfaces. They wrap the SQL
// repositories of each service directly so a test suite can exercise a
// service without booting its gRPC dependencies. These live here instead of
// in pkg/adapter so the shared pkg module never depends on service modules.

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	merchant_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/repository"
	saldo_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	user_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/jackc/pgx/v5/pgtype"
)

func tsPtr(t pgtype.Timestamp) *time.Time {
	if !t.Valid {
		return nil
	}
	tm := t.Time
	return &tm
}

func datePtr(d pgtype.Date) *time.Time {
	if !d.Valid {
		return nil
	}
	tm := d.Time
	return &tm
}

// --- user ---

type localUserAdapter struct {
	repo user_repo.UserQueryRepository
}

func NewLocalUserAdapter(repo user_repo.UserQueryRepository) adapter.UserAdapter {
	return &localUserAdapter{repo: repo}
}

func (a *localUserAdapter) FindById(ctx context.Context, userID int) (*models.User, error) {
	res, err := a.repo.FindById(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &models.User{
		UserID:    res.UserID,
		Email:     res.Email,
		Firstname: res.Firstname,
		Lastname:  res.Lastname,
	}, nil
}

// --- card ---

type localCardAdapter struct {
	queryRepo   card_repo.CardQueryRepository
	commandRepo card_repo.CardCommandRepository
}

func NewLocalCardAdapter(queryRepo card_repo.CardQueryRepository, commandRepo card_repo.CardCommandRepository) adapter.CardAdapter {
	return &localCardAdapter{queryRepo: queryRepo, commandRepo: commandRepo}
}

func mapCardFields(cardID, userID int32, cardNumber, cardType string, expireDate pgtype.Date, cvv, cardProvider string, createdAt, updatedAt pgtype.Timestamp) *models.Card {
	return &models.Card{
		CardID:       cardID,
		UserID:       userID,
		CardNumber:   cardNumber,
		CardType:     cardType,
		ExpireDate:   datePtr(expireDate),
		Cvv:          cvv,
		CardProvider: cardProvider,
		CreatedAt:    tsPtr(createdAt),
		UpdatedAt:    tsPtr(updatedAt),
	}
}

func (a *localCardAdapter) FindCardByUserId(ctx context.Context, user_id int) (*models.Card, error) {
	res, err := a.queryRepo.FindCardByUserId(ctx, user_id)
	if err != nil {
		return nil, err
	}
	return mapCardFields(res.CardID, res.UserID, res.CardNumber, res.CardType, res.ExpireDate, res.Cvv, res.CardProvider, res.CreatedAt, res.UpdatedAt), nil
}

func (a *localCardAdapter) FindUserCardByCardNumber(ctx context.Context, card_number string) (*models.Card, error) {
	res, err := a.queryRepo.FindUserCardByCardNumber(ctx, card_number)
	if err != nil {
		return nil, err
	}
	card := mapCardFields(res.CardID, res.UserID, res.CardNumber, res.CardType, res.ExpireDate, res.Cvv, res.CardProvider, res.CreatedAt, res.UpdatedAt)
	card.Email = res.Email
	return card, nil
}

func (a *localCardAdapter) FindCardByCardNumber(ctx context.Context, card_number string) (*models.Card, error) {
	res, err := a.queryRepo.FindCardByCardNumber(ctx, card_number)
	if err != nil {
		return nil, err
	}
	return mapCardFields(res.CardID, res.UserID, res.CardNumber, res.CardType, res.ExpireDate, res.Cvv, res.CardProvider, res.CreatedAt, res.UpdatedAt), nil
}

func (a *localCardAdapter) UpdateCard(ctx context.Context, request *requests.UpdateCardRequest) (*models.Card, error) {
	res, err := a.commandRepo.UpdateCard(ctx, request)
	if err != nil {
		return nil, err
	}
	return mapCardFields(res.CardID, res.UserID, res.CardNumber, res.CardType, res.ExpireDate, res.Cvv, res.CardProvider, res.CreatedAt, res.UpdatedAt), nil
}

// --- merchant ---

type localMerchantAdapter struct {
	repo merchant_repo.MerchantQueryRepository
}

func NewLocalMerchantAdapter(repo merchant_repo.MerchantQueryRepository) adapter.MerchantAdapter {
	return &localMerchantAdapter{repo: repo}
}

func (a *localMerchantAdapter) FindByApiKey(ctx context.Context, api_key string) (*models.Merchant, error) {
	res, err := a.repo.FindByApiKey(ctx, api_key)
	if err != nil {
		return nil, err
	}
	return &models.Merchant{
		MerchantID: res.MerchantID,
		Name:       res.Name,
		ApiKey:     res.ApiKey,
		UserID:     res.UserID,
		Status:     res.Status,
		CreatedAt:  tsPtr(res.CreatedAt),
		UpdatedAt:  tsPtr(res.UpdatedAt),
	}, nil
}

func (a *localMerchantAdapter) FindByMerchantId(ctx context.Context, merchant_id int) (*models.Merchant, error) {
	res, err := a.repo.FindByMerchantId(ctx, merchant_id)
	if err != nil {
		return nil, err
	}
	return &models.Merchant{
		MerchantID: res.MerchantID,
		Name:       res.Name,
		ApiKey:     res.ApiKey,
		UserID:     res.UserID,
		Status:     res.Status,
		CreatedAt:  tsPtr(res.CreatedAt),
		UpdatedAt:  tsPtr(res.UpdatedAt),
	}, nil
}

// --- saldo ---

type localSaldoAdapter struct {
	repo saldo_repo.Repositories
}

func NewLocalSaldoAdapter(repo saldo_repo.Repositories) adapter.SaldoAdapter {
	return &localSaldoAdapter{repo: repo}
}

func (a *localSaldoAdapter) FindByCardNumber(ctx context.Context, card_number string) (*models.Saldo, error) {
	res, err := a.repo.FindByCardNumber(ctx, card_number)
	if err != nil {
		return nil, err
	}
	return &models.Saldo{
		SaldoID:        res.SaldoID,
		CardNumber:     res.CardNumber,
		TotalBalance:   res.TotalBalance,
		WithdrawAmount: res.WithdrawAmount,
		WithdrawTime:   tsPtr(res.WithdrawTime),
		CreatedAt:      tsPtr(res.CreatedAt),
		UpdatedAt:      tsPtr(res.UpdatedAt),
	}, nil
}

func mapMutationRow(saldoID int32, cardNumber string, totalBalance int64) *models.SaldoMutationResult {
	return &models.SaldoMutationResult{
		SaldoID:      saldoID,
		CardNumber:   cardNumber,
		TotalBalance: totalBalance,
	}
}

func (a *localSaldoAdapter) UpdateSaldoBalance(ctx context.Context, request *requests.UpdateSaldoBalance) (*models.SaldoMutationResult, error) {
	res, err := a.repo.UpdateSaldoBalance(ctx, request)
	if err != nil {
		return nil, err
	}
	return mapMutationRow(res.SaldoID, res.CardNumber, res.TotalBalance), nil
}

func (a *localSaldoAdapter) DebitSaldo(ctx context.Context, request *requests.DebitSaldoRequest) (*models.SaldoMutationResult, error) {
	res, err := a.repo.DebitSaldo(ctx, request)
	if err != nil {
		return nil, err
	}
	return mapMutationRow(res.SaldoID, res.CardNumber, res.TotalBalance), nil
}

func (a *localSaldoAdapter) CreditSaldo(ctx context.Context, request *requests.CreditSaldoRequest) (*models.SaldoMutationResult, error) {
	res, err := a.repo.CreditSaldo(ctx, request)
	if err != nil {
		return nil, err
	}
	return mapMutationRow(res.SaldoID, res.CardNumber, res.TotalBalance), nil
}

func (a *localSaldoAdapter) UpdateSaldoWithdraw(ctx context.Context, request *requests.UpdateSaldoWithdraw) (*models.SaldoMutationResult, error) {
	res, err := a.repo.UpdateSaldoWithdraw(ctx, request)
	if err != nil {
		return nil, err
	}
	return mapMutationRow(res.SaldoID, res.CardNumber, res.TotalBalance), nil
}
