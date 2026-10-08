package repository

import (
	"context"

	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/jmoiron/sqlx"
)

// SQL source of truth (formerly database/query/withdraw.sql).
const (
	createWithdraw = `INSERT INTO
    withdraws (
        card_number,
        withdraw_amount,
        withdraw_time,
        created_at,
        updated_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        current_timestamp
    )
RETURNING
    withdraw_id,
    withdraw_no,
    card_number,
    withdraw_amount,
    withdraw_time,
    status,
    created_at,
    updated_at`

	updateWithdraw = `UPDATE withdraws
SET
    card_number = $2,
    withdraw_amount = $3,
    withdraw_time = $4,
    updated_at = current_timestamp
WHERE
    withdraw_id = $1
    AND deleted_at IS NULL
RETURNING
    withdraw_id,
    withdraw_no,
    card_number,
    withdraw_amount,
    withdraw_time,
    status,
    created_at,
    updated_at`

	updateWithdrawStatus = `UPDATE withdraws
SET
    status = $2,
    updated_at = current_timestamp
WHERE
    withdraw_id = $1
    AND deleted_at IS NULL
RETURNING
    withdraw_id,
    withdraw_no,
    card_number,
    withdraw_amount,
    withdraw_time,
    status,
    created_at,
    updated_at`

	trashWithdraw = `UPDATE withdraws
SET
    deleted_at = current_timestamp
WHERE
    withdraw_id = $1
    AND deleted_at IS NULL
RETURNING
    withdraw_id,
    withdraw_no,
    card_number,
    withdraw_amount,
    withdraw_time,
    status,
    created_at,
    updated_at,
    deleted_at`

	restoreWithdraw = `UPDATE withdraws
SET
    deleted_at = NULL
WHERE
    withdraw_id = $1
    AND deleted_at IS NOT NULL
RETURNING
    withdraw_id,
    withdraw_no,
    card_number,
    withdraw_amount,
    withdraw_time,
    status,
    created_at,
    updated_at,
    deleted_at`

	deleteWithdrawPermanently = `DELETE FROM withdraws
WHERE
    withdraw_id = $1
    AND deleted_at IS NOT NULL`

	restoreAllWithdraws = `UPDATE withdraws
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL`

	deleteAllPermanentWithdraws = `DELETE FROM withdraws WHERE deleted_at IS NOT NULL`
)

type withdrawCommandRepository struct {
	db *sqlx.DB
}

func NewWithdrawCommandRepository(db *sqlx.DB) WithdrawCommandRepository {
	return &withdrawCommandRepository{
		db: db,
	}
}

func (r *withdrawCommandRepository) CreateWithdraw(ctx context.Context, request *requests.CreateWithdrawRequest) (*db.CreateWithdrawRow, error) {
	var row db.CreateWithdrawRow
	err := r.db.GetContext(ctx, &row, createWithdraw,
		request.CardNumber,
		int64(request.WithdrawAmount),
		request.WithdrawTime,
		request.WithdrawTime,
	)

	if err != nil {
		return nil, sharedErrors.ErrFailed("create withdraw").WithInternal(err)
	}

	return &row, nil
}

func (r *withdrawCommandRepository) UpdateWithdraw(ctx context.Context, request *requests.UpdateWithdrawRequest) (*db.UpdateWithdrawRow, error) {
	var row db.UpdateWithdrawRow
	err := r.db.GetContext(ctx, &row, updateWithdraw,
		int32(*request.WithdrawID),
		request.CardNumber,
		int64(request.WithdrawAmount),
		request.WithdrawTime,
	)

	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "withdraw", "update withdraw")
	}

	return &row, nil
}

func (r *withdrawCommandRepository) UpdateWithdrawStatus(ctx context.Context, request *requests.UpdateWithdrawStatus) (*db.UpdateWithdrawStatusRow, error) {
	var row db.UpdateWithdrawStatusRow
	err := r.db.GetContext(ctx, &row, updateWithdrawStatus,
		int32(request.WithdrawID),
		request.Status,
	)

	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "withdraw", "update withdraw status")
	}

	return &row, nil
}

func (r *withdrawCommandRepository) TrashedWithdraw(ctx context.Context, withdraw_id int) (*db.Withdraw, error) {
	var row db.Withdraw
	err := r.db.GetContext(ctx, &row, trashWithdraw, int32(withdraw_id))

	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "withdraw", "trash withdraw")
	}

	return &row, nil
}

func (r *withdrawCommandRepository) RestoreWithdraw(ctx context.Context, withdraw_id int) (*db.Withdraw, error) {
	var row db.Withdraw
	err := r.db.GetContext(ctx, &row, restoreWithdraw, int32(withdraw_id))

	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "withdraw", "restore withdraw")
	}

	return &row, nil
}

func (r *withdrawCommandRepository) DeleteWithdrawPermanent(ctx context.Context, withdraw_id int) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteWithdrawPermanently, int32(withdraw_id)); err != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(err, "withdraw", "delete withdraw permanently")
	}

	return true, nil
}

func (r *withdrawCommandRepository) RestoreAllWithdraw(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, restoreAllWithdraws); err != nil {
		return false, sharedErrors.ErrFailed("restore all withdraws").WithInternal(err)
	}

	return true, nil
}

func (r *withdrawCommandRepository) DeleteAllWithdrawPermanent(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteAllPermanentWithdraws); err != nil {
		return false, sharedErrors.ErrFailed("delete all withdraws permanently").WithInternal(err)
	}

	return true, nil
}
