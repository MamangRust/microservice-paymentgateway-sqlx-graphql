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

// merchantDocumentQueryRepository provides methods to query merchant documents from the database.
type merchantDocumentQueryRepository struct {
	db *sqlx.DB
}

// NewMerchantDocumentQueryRepository creates a new instance of merchantDocumentQueryRepository.
func NewMerchantDocumentQueryRepository(db *sqlx.DB) MerchantDocumentQueryRepository {
	return &merchantDocumentQueryRepository{
		db: db,
	}
}

func (r *merchantDocumentQueryRepository) FindAllDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*db.GetMerchantDocumentsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var docs []*db.GetMerchantDocumentsRow
	if err := r.db.SelectContext(ctx, &docs, getMerchantDocuments,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	); err != nil {
		return nil, sharedErrors.ErrFailed("find all merchant documents").WithInternal(err)
	}

	return docs, nil
}

func (r *merchantDocumentQueryRepository) FindByActiveDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*db.GetActiveMerchantDocumentsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var docs []*db.GetActiveMerchantDocumentsRow
	if err := r.db.SelectContext(ctx, &docs, getActiveMerchantDocuments,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	); err != nil {
		return nil, sharedErrors.ErrFailed("find active merchant documents").WithInternal(err)
	}

	return docs, nil
}

func (r *merchantDocumentQueryRepository) FindByTrashedDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*db.GetTrashedMerchantDocumentsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var docs []*db.GetTrashedMerchantDocumentsRow
	if err := r.db.SelectContext(ctx, &docs, getTrashedMerchantDocuments,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	); err != nil {
		return nil, sharedErrors.ErrFailed("find trashed merchant documents").WithInternal(err)
	}

	return docs, nil
}

func (r *merchantDocumentQueryRepository) FindByIdDocument(ctx context.Context, id int) (*db.GetMerchantDocumentRow, error) {
	var doc db.GetMerchantDocumentRow
	if err := r.db.GetContext(ctx, &doc, getMerchantDocument, int32(id)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFoundResponse("merchant document").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &doc, nil
}
