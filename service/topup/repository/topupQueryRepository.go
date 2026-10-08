package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/topup/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/jmoiron/sqlx"
)

const getTopups = `
SELECT
    topup_id,
    topup_no,
    card_number,
    topup_amount,
    topup_method,
    topup_time,
    status,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM topups
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR card_number ILIKE '%' || $1 || '%'
        OR topup_no::TEXT ILIKE '%' || $1 || '%'
        OR topup_method ILIKE '%' || $1 || '%'
        OR status ILIKE '%' || $1 || '%'
    )
ORDER BY topup_time DESC
LIMIT $2
OFFSET
    $3
`

const getActiveTopups = `
SELECT
    topup_id,
    topup_no,
    card_number,
    topup_amount,
    topup_method,
    topup_time,
    status,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM topups
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR card_number ILIKE '%' || $1 || '%'
        OR topup_no::TEXT ILIKE '%' || $1 || '%'
        OR topup_method ILIKE '%' || $1 || '%'
    )
ORDER BY topup_time DESC
LIMIT $2
OFFSET
    $3
`

const getTrashedTopups = `
SELECT
    topup_id,
    topup_no,
    card_number,
    topup_amount,
    topup_method,
    topup_time,
    status,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM topups
WHERE
    deleted_at IS NOT NULL
    AND (
        $1::TEXT IS NULL
        OR card_number ILIKE '%' || $1 || '%'
        OR topup_no::TEXT ILIKE '%' || $1 || '%'
        OR topup_method ILIKE '%' || $1 || '%'
    )
ORDER BY topup_time DESC
LIMIT $2
OFFSET
    $3
`

const getTopupsByCardNumber = `
SELECT
    topup_id,
    topup_no,
    card_number,
    topup_amount,
    topup_method,
    topup_time,
    status,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM topups
WHERE
    deleted_at IS NULL
    AND card_number = $1
    AND (
        $2::TEXT IS NULL
        OR topup_no::TEXT ILIKE '%' || $2 || '%'
        OR topup_method ILIKE '%' || $2 || '%'
        OR status ILIKE '%' || $2 || '%'
    )
ORDER BY topup_time DESC
LIMIT $3
OFFSET
    $4
`

const getTopupByID = `
SELECT
    topup_id,
    topup_no,
    card_number,
    topup_amount,
    topup_method,
    topup_time,
    status,
    created_at,
    updated_at
FROM topups
WHERE
    topup_id = $1
    AND deleted_at IS NULL
`

type topupQueryRepository struct {
	db *sqlx.DB
}

func NewTopupQueryRepository(db *sqlx.DB) TopupQueryRepository {
	return &topupQueryRepository{
		db: db,
	}
}

func (r *topupQueryRepository) FindAllTopups(ctx context.Context, req *requests.FindAllTopups) ([]*db.GetTopupsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetTopupsRow

	err := r.db.SelectContext(ctx, &res, getTopups,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, sharedErrors.ErrFailed("find all topups").WithInternal(err)
	}

	return res, nil
}

func (r *topupQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllTopups) ([]*db.GetActiveTopupsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetActiveTopupsRow

	err := r.db.SelectContext(ctx, &res, getActiveTopups,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, sharedErrors.ErrFailed("find active topups").WithInternal(err)
	}

	return res, nil
}

func (r *topupQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllTopups) ([]*db.GetTrashedTopupsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetTrashedTopupsRow

	err := r.db.SelectContext(ctx, &res, getTrashedTopups,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, sharedErrors.ErrFailed("find trashed topups").WithInternal(err)
	}

	return res, nil
}

func (r *topupQueryRepository) FindAllTopupByCardNumber(ctx context.Context, req *requests.FindAllTopupsByCardNumber) ([]*db.GetTopupsByCardNumberRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetTopupsByCardNumberRow

	err := r.db.SelectContext(ctx, &res, getTopupsByCardNumber,
		req.CardNumber,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, sharedErrors.ErrFailed("find topups by card number").WithInternal(err)
	}

	return res, nil
}

func (r *topupQueryRepository) FindById(ctx context.Context, topup_id int) (*db.GetTopupByIDRow, error) {
	var res db.GetTopupByIDRow

	err := r.db.GetContext(ctx, &res, getTopupByID, int32(topup_id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFoundResponse("topup").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &res, nil
}
