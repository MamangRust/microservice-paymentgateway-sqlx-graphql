package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/jmoiron/sqlx"
)

type transferQueryRepository struct {
	db *sqlx.DB
}

func NewTransferQueryRepository(db *sqlx.DB) TransferQueryRepository {
	return &transferQueryRepository{
		db: db,
	}
}

func (r *transferQueryRepository) FindAll(ctx context.Context, req *requests.FindAllTransfers) ([]*db.GetTransfersRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetTransfersRow
	err := r.db.SelectContext(ctx, &res, getTransfers, req.Search, int32(req.PageSize), int32(offset))

	if err != nil {
		return nil, sharedErrors.ErrFailed("find all transfers").WithInternal(err)
	}

	return res, nil
}

func (r *transferQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllTransfers) ([]*db.GetActiveTransfersRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetActiveTransfersRow
	err := r.db.SelectContext(ctx, &res, getActiveTransfers, req.Search, int32(req.PageSize), int32(offset))

	if err != nil {
		return nil, sharedErrors.ErrFailed("find active transfers").WithInternal(err)
	}

	return res, nil
}

func (r *transferQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllTransfers) ([]*db.GetTrashedTransfersRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetTrashedTransfersRow
	err := r.db.SelectContext(ctx, &res, getTrashedTransfers, req.Search, int32(req.PageSize), int32(offset))

	if err != nil {
		return nil, sharedErrors.ErrFailed("find trashed transfers").WithInternal(err)
	}

	return res, nil
}

func (r *transferQueryRepository) FindById(ctx context.Context, id int) (*db.GetTransferByIDRow, error) {
	var transfer db.GetTransferByIDRow
	err := r.db.GetContext(ctx, &transfer, getTransferByID, int32(id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFoundResponse("transfer").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &transfer, nil
}

func (r *transferQueryRepository) FindTransferByTransferFrom(ctx context.Context, transfer_from string) ([]*db.GetTransfersBySourceCardRow, error) {
	var res []*db.GetTransfersBySourceCardRow
	err := r.db.SelectContext(ctx, &res, getTransfersBySourceCard, transfer_from)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFoundResponse("transfer").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return res, nil
}

func (r *transferQueryRepository) FindTransferByTransferTo(ctx context.Context, transfer_to string) ([]*db.GetTransfersByDestinationCardRow, error) {
	var res []*db.GetTransfersByDestinationCardRow
	err := r.db.SelectContext(ctx, &res, getTransfersByDestinationCard, transfer_to)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFoundResponse("transfer").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return res, nil
}
