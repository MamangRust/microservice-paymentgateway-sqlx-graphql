package repository

import (
	"context"
	"time"

	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/topup/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/jmoiron/sqlx"
)

const createTopup = `
INSERT INTO
    topups (
        card_number,
        topup_amount,
        topup_method,
        topup_time,
        created_at,
        updated_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        current_timestamp,
        current_timestamp
    )
RETURNING
    topup_id,
    topup_no,
    card_number,
    topup_amount,
    topup_method,
    topup_time,
    status,
    created_at,
    updated_at
`

const deleteAllPermanentTopups = `
DELETE FROM topups WHERE deleted_at IS NOT NULL
`

const deleteTopupPermanently = `
DELETE FROM topups WHERE topup_id = $1
`

const restoreAllTopups = `
UPDATE topups
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL
`

const restoreTopup = `
UPDATE topups
SET
    deleted_at = NULL
WHERE
    topup_id = $1
    AND deleted_at IS NOT NULL
RETURNING
    topup_id,
    topup_no,
    card_number,
    topup_amount,
    topup_method,
    topup_time,
    status,
    created_at,
    updated_at,
    deleted_at
`

const trashTopup = `
UPDATE topups
SET
    deleted_at = current_timestamp
WHERE
    topup_id = $1
    AND deleted_at IS NULL
RETURNING
    topup_id,
    topup_no,
    card_number,
    topup_amount,
    topup_method,
    topup_time,
    status,
    created_at,
    updated_at,
    deleted_at
`

const updateTopup = `
UPDATE topups
SET
    card_number = $2,
    topup_amount = $3,
    topup_method = $4,
    topup_time = $5,
    updated_at = current_timestamp
WHERE
    topup_id = $1
    AND deleted_at IS NULL
RETURNING
    topup_id,
    topup_no,
    card_number,
    topup_amount,
    topup_method,
    topup_time,
    status,
    created_at,
    updated_at
`

const updateTopupAmount = `
UPDATE topups
SET
    topup_amount = $2,
    updated_at = current_timestamp
WHERE
    topup_id = $1
    AND deleted_at IS NULL
RETURNING
    topup_id,
    topup_no,
    card_number,
    topup_amount,
    topup_method,
    topup_time,
    status,
    created_at,
    updated_at
`

const updateTopupStatus = `
UPDATE topups
SET
    status = $2,
    updated_at = current_timestamp
WHERE
    topup_id = $1
    AND deleted_at IS NULL
RETURNING
    topup_id,
    topup_no,
    card_number,
    topup_amount,
    topup_method,
    topup_time,
    status,
    created_at,
    updated_at
`

type topupCommandRepository struct {
	db *sqlx.DB
}

func NewTopupCommandRepository(db *sqlx.DB) TopupCommandRepository {
	return &topupCommandRepository{
		db: db,
	}
}

func (r *topupCommandRepository) CreateTopup(ctx context.Context, request *requests.CreateTopupRequest) (*db.CreateTopupRow, error) {
	var res db.CreateTopupRow

	err := r.db.GetContext(ctx, &res, createTopup,
		request.CardNumber,
		int64(request.TopupAmount),
		request.TopupMethod,
		time.Time{},
	)

	if err != nil {
		return nil, sharedErrors.ErrFailed("create topup").WithInternal(err)
	}

	return &res, nil
}

func (r *topupCommandRepository) UpdateTopup(ctx context.Context, request *requests.UpdateTopupRequest) (*db.UpdateTopupRow, error) {
	var res db.UpdateTopupRow

	err := r.db.GetContext(ctx, &res, updateTopup,
		int32(*request.TopupID),
		request.CardNumber,
		int64(request.TopupAmount),
		request.TopupMethod,
		time.Time{},
	)

	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "topup", "update topup")
	}

	return &res, nil
}

func (r *topupCommandRepository) UpdateTopupAmount(ctx context.Context, request *requests.UpdateTopupAmount) (*db.UpdateTopupAmountRow, error) {
	var res db.UpdateTopupAmountRow

	err := r.db.GetContext(ctx, &res, updateTopupAmount,
		int32(request.TopupID),
		int64(request.TopupAmount),
	)

	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "topup", "update topup amount")
	}

	return &res, nil
}

func (r *topupCommandRepository) UpdateTopupStatus(ctx context.Context, request *requests.UpdateTopupStatus) (*db.UpdateTopupStatusRow, error) {
	var res db.UpdateTopupStatusRow

	err := r.db.GetContext(ctx, &res, updateTopupStatus,
		int32(request.TopupID),
		request.Status,
	)

	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "topup", "update topup status")
	}

	return &res, nil
}

func (r *topupCommandRepository) TrashedTopup(ctx context.Context, topup_id int) (*db.Topup, error) {
	var res db.Topup

	err := r.db.GetContext(ctx, &res, trashTopup, int32(topup_id))
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "topup", "trash topup")
	}
	return &res, nil
}

func (r *topupCommandRepository) RestoreTopup(ctx context.Context, topup_id int) (*db.Topup, error) {
	var res db.Topup

	err := r.db.GetContext(ctx, &res, restoreTopup, int32(topup_id))
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "topup", "restore topup")
	}
	return &res, nil
}

func (r *topupCommandRepository) DeleteTopupPermanent(ctx context.Context, topup_id int) (bool, error) {
	_, err := r.db.ExecContext(ctx, deleteTopupPermanently, int32(topup_id))
	if err != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(err, "topup", "delete topup permanently")
	}
	return true, nil
}

func (r *topupCommandRepository) RestoreAllTopup(ctx context.Context) (bool, error) {
	_, err := r.db.ExecContext(ctx, restoreAllTopups)

	if err != nil {
		return false, sharedErrors.ErrFailed("restore all topups").WithInternal(err)
	}

	return true, nil
}

func (r *topupCommandRepository) DeleteAllTopupPermanent(ctx context.Context) (bool, error) {
	_, err := r.db.ExecContext(ctx, deleteAllPermanentTopups)

	if err != nil {
		return false, sharedErrors.ErrFailed("delete all topups permanently").WithInternal(err)
	}

	return true, nil
}
