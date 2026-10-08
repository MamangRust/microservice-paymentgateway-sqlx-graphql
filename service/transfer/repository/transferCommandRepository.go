package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"

	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/state"
	"github.com/jmoiron/sqlx"
)

type transferCommandRepository struct {
	db *sqlx.DB
}

func NewTransferCommandRepository(db *sqlx.DB) TransferCommandRepository {
	return &transferCommandRepository{
		db: db,
	}
}

func (r *transferCommandRepository) CreateTransfer(ctx context.Context, request *requests.CreateTransferRequest) (*db.CreateTransferRow, error) {
	var row db.CreateTransferRow
	err := r.db.GetContext(ctx, &row, createTransfer,
		request.TransferFrom,
		request.TransferTo,
		int64(request.TransferAmount),
		time.Now(),
		state.Pending,
	)

	if err != nil {
		return nil, sharedErrors.ErrFailed("create transfer").WithInternal(err)
	}

	return &row, nil
}

func (r *transferCommandRepository) UpdateTransfer(ctx context.Context, request *requests.UpdateTransferRequest) (*db.UpdateTransferRow, error) {
	var row db.UpdateTransferRow
	err := r.db.GetContext(ctx, &row, updateTransfer,
		int32(*request.TransferID),
		request.TransferFrom,
		request.TransferTo,
		int64(request.TransferAmount),
		time.Now(),
	)

	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transfer", "update transfer")
	}

	return &row, nil

}

func (r *transferCommandRepository) UpdateTransferAmount(ctx context.Context, request *requests.UpdateTransferAmountRequest) (*db.UpdateTransferAmountRow, error) {
	var row db.UpdateTransferAmountRow
	err := r.db.GetContext(ctx, &row, updateTransferAmount,
		int32(request.TransferID),
		int64(request.TransferAmount),
	)

	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transfer", "update transfer amount")
	}

	return &row, nil
}

func (r *transferCommandRepository) UpdateTransferStatus(ctx context.Context, request *requests.UpdateTransferStatus) (*db.UpdateTransferStatusRow, error) {
	var row db.UpdateTransferStatusRow
	err := r.db.GetContext(ctx, &row, updateTransferStatus,
		int32(request.TransferID),
		request.Status,
	)

	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transfer", "update transfer status")
	}

	return &row, nil
}

// GuardStatus atomically transitions a transfer from an expected status to a new
// one. It reports false (with nil error) when the row is not in the expected
// status, i.e. another actor already moved it.
func (r *transferCommandRepository) GuardStatus(ctx context.Context, id int, fromStatus, toStatus, reason string) (bool, error) {
	if err := state.CheckTransition(fromStatus, toStatus); err != nil {
		return false, err
	}
	var row db.UpdateTransferStatusRow
	err := r.db.GetContext(ctx, &row, guardTransferStatus, int32(id), fromStatus, toStatus, reason)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ListStuck returns transfers stuck in a recoverable state for the recovery worker.
func (r *transferCommandRepository) TransitionStatus(ctx context.Context, id int, fromStatus, toStatus, reason string) (*db.UpdateTransferStatusRow, error) {
	if err := state.CheckTransition(fromStatus, toStatus); err != nil {
		return nil, err
	}
	var row db.UpdateTransferStatusRow
	err := r.db.GetContext(ctx, &row, guardTransferStatus, int32(id), fromStatus, toStatus, reason)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *transferCommandRepository) ListStuck(ctx context.Context, olderThan time.Duration, maxRows int32) ([]*db.StuckTransfer, error) {
	var res []*db.StuckTransfer
	err := r.db.SelectContext(ctx, &res, listStuckTransfers, int64(olderThan.Seconds()), maxRows)
	return res, err
}

func (r *transferCommandRepository) TrashedTransfer(ctx context.Context, transfer_id int) (*db.Transfer, error) {
	var row db.Transfer
	err := r.db.GetContext(ctx, &row, trashTransfer, int32(transfer_id))

	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transfer", "trash transfer")
	}
	return &row, nil
}

func (r *transferCommandRepository) RestoreTransfer(ctx context.Context, transfer_id int) (*db.Transfer, error) {
	var row db.Transfer
	err := r.db.GetContext(ctx, &row, restoreTransfer, int32(transfer_id))

	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "transfer", "restore transfer")
	}
	return &row, nil
}

func (r *transferCommandRepository) DeleteTransferPermanent(ctx context.Context, transfer_id int) (bool, error) {
	_, err := r.db.ExecContext(ctx, deleteTransferPermanently, int32(transfer_id))
	if err != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(err, "transfer", "delete transfer permanently")
	}
	return true, nil
}

func (r *transferCommandRepository) RestoreAllTransfer(ctx context.Context) (bool, error) {
	_, err := r.db.ExecContext(ctx, restoreAllTransfers)

	if err != nil {
		return false, sharedErrors.ErrFailed("restore all transfers").WithInternal(err)
	}

	return true, nil
}

func (r *transferCommandRepository) DeleteAllTransferPermanent(ctx context.Context) (bool, error) {
	_, err := r.db.ExecContext(ctx, deleteAllPermanentTransfers)

	if err != nil {
		return false, sharedErrors.ErrFailed("delete all transfers permanently").WithInternal(err)
	}

	return true, nil
}
