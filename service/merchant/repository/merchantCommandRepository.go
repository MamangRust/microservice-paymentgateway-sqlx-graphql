package repository

import (
	"context"

	apikey "github.com/MamangRust/microservice-payment-gateway-grpc/pkg/api-key"
	database "github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database"
	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/jmoiron/sqlx"
)

type merchantCommandRepository struct {
	db *sqlx.DB
}

func NewMerchantCommandRepository(db *sqlx.DB) MerchantCommandRepository {
	return &merchantCommandRepository{
		db: db,
	}
}

func (r *merchantCommandRepository) CreateMerchant(ctx context.Context, request *requests.CreateMerchantRequest) (*db.CreateMerchantRow, error) {
	apiKey, err := apikey.GenerateApiKey()

	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	var merchant db.CreateMerchantRow
	if err := r.db.GetContext(ctx, &merchant, createMerchant,
		request.Name,
		apiKey,
		int32(request.UserID),
		"inactive",
	); err != nil {
		if database.IsUniqueViolation(err) {
			return nil, sharedErrors.NewConflictError("merchant api key already exists").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("create merchant").WithInternal(err)
	}

	return &merchant, nil
}

func (r *merchantCommandRepository) UpdateMerchant(ctx context.Context, request *requests.UpdateMerchantRequest) (*db.UpdateMerchantRow, error) {
	var merchant db.UpdateMerchantRow
	if err := r.db.GetContext(ctx, &merchant, updateMerchant,
		int32(*request.MerchantID),
		request.Name,
		int32(request.UserID),
		request.Status,
	); err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "merchant", "update merchant")
	}

	return &merchant, nil
}

func (r *merchantCommandRepository) UpdateMerchantStatus(ctx context.Context, request *requests.UpdateMerchantStatusRequest) (*db.UpdateMerchantStatusRow, error) {
	var merchant db.UpdateMerchantStatusRow
	if err := r.db.GetContext(ctx, &merchant, updateMerchantStatus,
		int32(*request.MerchantID),
		request.Status,
	); err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "merchant", "update merchant status")
	}

	return &merchant, nil
}

func (r *merchantCommandRepository) TrashedMerchant(ctx context.Context, merchant_id int) (*db.Merchant, error) {
	var merchant db.Merchant
	if err := r.db.GetContext(ctx, &merchant, trashMerchant, int32(merchant_id)); err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "merchant", "trash merchant")
	}

	return &merchant, nil
}

func (r *merchantCommandRepository) RestoreMerchant(ctx context.Context, merchant_id int) (*db.Merchant, error) {
	var merchant db.Merchant
	if err := r.db.GetContext(ctx, &merchant, restoreMerchant, int32(merchant_id)); err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "merchant", "restore merchant")
	}

	return &merchant, nil
}

func (r *merchantCommandRepository) DeleteMerchantPermanent(ctx context.Context, merchant_id int) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteMerchantPermanently, int32(merchant_id)); err != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(err, "merchant", "delete merchant permanently")
	}

	return true, nil
}

func (r *merchantCommandRepository) RestoreAllMerchant(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, restoreAllMerchants); err != nil {
		return false, sharedErrors.ErrFailed("restore all merchants").WithInternal(err)
	}

	return true, nil
}

func (r *merchantCommandRepository) DeleteAllMerchantPermanent(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteAllPermanentMerchants); err != nil {
		return false, sharedErrors.ErrFailed("delete all merchants permanently").WithInternal(err)
	}

	return true, nil
}
