package repository

import (
	"context"

	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/jmoiron/sqlx"
)

type merchantDocumentCommandRepository struct {
	db *sqlx.DB
}

func NewMerchantDocumentCommandRepository(db *sqlx.DB) MerchantDocumentCommandRepository {
	return &merchantDocumentCommandRepository{
		db: db,
	}
}

func (r *merchantDocumentCommandRepository) CreateMerchantDocument(ctx context.Context, request *requests.CreateMerchantDocumentRequest) (*db.CreateMerchantDocumentRow, error) {
	note := ""

	var doc db.CreateMerchantDocumentRow
	if err := r.db.GetContext(ctx, &doc, createMerchantDocument,
		int32(request.MerchantID),
		request.DocumentType,
		request.DocumentUrl,
		"pending",
		&note,
	); err != nil {
		return nil, sharedErrors.ErrFailed("create merchant document").WithInternal(err)
	}

	return &doc, nil
}

func (r *merchantDocumentCommandRepository) UpdateMerchantDocument(ctx context.Context, request *requests.UpdateMerchantDocumentRequest) (*db.UpdateMerchantDocumentRow, error) {
	if request.DocumentID == nil || *request.DocumentID <= 0 {
		return nil, sharedErrors.NewBadRequestError("merchant document ID is required")
	}

	note := ""

	var doc db.UpdateMerchantDocumentRow
	if err := r.db.GetContext(ctx, &doc, updateMerchantDocument,
		int32(*request.DocumentID),
		request.DocumentType,
		request.DocumentUrl,
		request.Status,
		&note,
	); err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "merchant document", "update merchant document")
	}

	return &doc, nil
}

func (r *merchantDocumentCommandRepository) UpdateMerchantDocumentStatus(ctx context.Context, request *requests.UpdateMerchantDocumentStatusRequest) (*db.UpdateMerchantDocumentStatusRow, error) {
	if request.DocumentID == nil || *request.DocumentID <= 0 {
		return nil, sharedErrors.NewBadRequestError("merchant document ID is required")
	}

	note := ""

	var doc db.UpdateMerchantDocumentStatusRow
	if err := r.db.GetContext(ctx, &doc, updateMerchantDocumentStatus,
		int32(*request.DocumentID),
		request.Status,
		&note,
	); err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "merchant document", "update merchant document status")
	}

	return &doc, nil
}

func (r *merchantDocumentCommandRepository) TrashedMerchantDocument(ctx context.Context, documentID int) (*db.MerchantDocument, error) {
	var doc db.MerchantDocument
	if err := r.db.GetContext(ctx, &doc, trashMerchantDocument, int32(documentID)); err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "merchant document", "trash merchant document")
	}

	return &doc, nil
}

func (r *merchantDocumentCommandRepository) RestoreMerchantDocument(ctx context.Context, documentID int) (*db.MerchantDocument, error) {
	var doc db.MerchantDocument
	if err := r.db.GetContext(ctx, &doc, restoreMerchantDocument, int32(documentID)); err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "merchant document", "restore merchant document")
	}

	return &doc, nil
}

func (r *merchantDocumentCommandRepository) DeleteMerchantDocumentPermanent(ctx context.Context, documentID int) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteMerchantDocumentPermanently, int32(documentID)); err != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(err, "merchant document", "delete merchant document permanently")
	}

	return true, nil
}

func (r *merchantDocumentCommandRepository) RestoreAllMerchantDocument(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, restoreAllMerchantDocuments); err != nil {
		return false, sharedErrors.ErrFailed("restore all merchant documents").WithInternal(err)
	}

	return true, nil
}

func (r *merchantDocumentCommandRepository) DeleteAllMerchantDocumentPermanent(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteAllPermanentMerchantDocuments); err != nil {
		return false, sharedErrors.ErrFailed("delete all merchant documents permanently").WithInternal(err)
	}

	return true, nil
}
