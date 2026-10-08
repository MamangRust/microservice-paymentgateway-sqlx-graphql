package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/jmoiron/sqlx"
)

type merchantQueryRepository struct {
	db *sqlx.DB
}

func NewMerchantQueryRepository(db *sqlx.DB) MerchantQueryRepository {
	return &merchantQueryRepository{
		db: db,
	}
}

func (r *merchantQueryRepository) FindAllMerchants(ctx context.Context, req *requests.FindAllMerchants) ([]*db.GetMerchantsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var merchants []*db.GetMerchantsRow
	err := r.db.SelectContext(ctx, &merchants, getMerchants,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, sharedErrors.ErrFailed("find all merchants").WithInternal(err)
	}

	return merchants, nil
}

func (r *merchantQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllMerchants) ([]*db.GetActiveMerchantsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var merchants []*db.GetActiveMerchantsRow
	err := r.db.SelectContext(ctx, &merchants, getActiveMerchants,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, sharedErrors.ErrFailed("find active merchants").WithInternal(err)
	}

	return merchants, nil
}

func (r *merchantQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllMerchants) ([]*db.GetTrashedMerchantsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var merchants []*db.GetTrashedMerchantsRow
	err := r.db.SelectContext(ctx, &merchants, getTrashedMerchants,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, sharedErrors.ErrFailed("find trashed merchants").WithInternal(err)
	}

	return merchants, nil
}

func (r *merchantQueryRepository) FindByMerchantId(ctx context.Context, merchant_id int) (*db.GetMerchantByIDRow, error) {
	var merchant db.GetMerchantByIDRow
	if err := r.db.GetContext(ctx, &merchant, getMerchantByID, int32(merchant_id)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFoundResponse("merchant").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &merchant, nil
}

func (r *merchantQueryRepository) FindByApiKey(ctx context.Context, api_key string) (*db.GetMerchantByApiKeyRow, error) {
	var merchant db.GetMerchantByApiKeyRow
	if err := r.db.GetContext(ctx, &merchant, getMerchantByApiKey, api_key); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFoundResponse("merchant").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &merchant, nil
}

func (r *merchantQueryRepository) FindByName(ctx context.Context, name string) (*db.GetMerchantByNameRow, error) {
	var merchant db.GetMerchantByNameRow
	if err := r.db.GetContext(ctx, &merchant, getMerchantByName, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFoundResponse("merchant").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &merchant, nil
}

func (r *merchantQueryRepository) FindByMerchantUserId(ctx context.Context, user_id int) ([]*db.GetMerchantsByUserIDRow, error) {
	var merchants []*db.GetMerchantsByUserIDRow
	err := r.db.SelectContext(ctx, &merchants, getMerchantsByUserID, int32(user_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFoundResponse("merchant").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return merchants, nil
}
