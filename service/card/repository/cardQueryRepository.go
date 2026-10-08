package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
)

// GetCards: Retrieves paginated list of active cards with search capability.
const getCards = `SELECT
    card_id,
    user_id,
    card_number,
    card_type,
    expire_date,
    cvv,
    card_provider,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM cards
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR card_number ILIKE '%' || $1 || '%'
        OR card_type ILIKE '%' || $1 || '%'
        OR card_provider ILIKE '%' || $1 || '%'
    )
ORDER BY card_id
LIMIT $2
OFFSET
    $3
`

const getActiveCardsWithCount = `SELECT
    card_id,
    user_id,
    card_number,
    card_type,
    expire_date,
    cvv,
    card_provider,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM cards
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR card_number ILIKE '%' || $1 || '%'
        OR card_type ILIKE '%' || $1 || '%'
        OR card_provider ILIKE '%' || $1 || '%'
    )
ORDER BY card_id
LIMIT $2
OFFSET
    $3
`

const getTrashedCardsWithCount = `SELECT
    card_id,
    user_id,
    card_number,
    card_type,
    expire_date,
    cvv,
    card_provider,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM cards
WHERE
    deleted_at IS NOT NULL
    AND (
        $1::TEXT IS NULL
        OR card_number ILIKE '%' || $1 || '%'
        OR card_type ILIKE '%' || $1 || '%'
        OR card_provider ILIKE '%' || $1 || '%'
    )
ORDER BY card_id
LIMIT $2
OFFSET
    $3
`

const getCardByID = `SELECT
    card_id,
    user_id,
    card_number,
    card_type,
    expire_date,
    cvv,
    card_provider,
    created_at,
    updated_at
FROM cards
WHERE
    card_id = $1
    AND deleted_at IS NULL
`

const getCardByUserID = `SELECT
    card_id,
    user_id,
    card_number,
    card_type,
    expire_date,
    cvv,
    card_provider,
    created_at,
    updated_at
FROM cards
WHERE
    user_id = $1
    AND deleted_at IS NULL
LIMIT 1
`

const getCardByCardNumber = `SELECT
    card_id,
    user_id,
    card_number,
    card_type,
    expire_date,
    cvv,
    card_provider,
    created_at,
    updated_at
FROM cards
WHERE
    card_number = $1
    AND deleted_at IS NULL
`

// GetUserEmailByCardNumber retrieves an active card by number.
// The card database owns cards, while user profile data lives in user_db.
// Email is enriched by the card service through UserQueryService; do not join users here.
const getUserEmailByCardNumber = `SELECT
    card_id,
    ''::TEXT AS email,
    user_id,
    card_number,
    card_type,
    expire_date,
    cvv,
    card_provider,
    created_at,
    updated_at
FROM cards
WHERE
    card_number = $1
    AND deleted_at IS NULL
`

type cardQueryRepository struct {
	db *sqlx.DB
}

func NewCardQueryRepository(db *sqlx.DB) CardQueryRepository {
	return &cardQueryRepository{
		db: db,
	}
}

func (r *cardQueryRepository) FindAllCards(ctx context.Context, req *requests.FindAllCards) ([]*db.GetCardsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var cards []*db.GetCardsRow
	err := r.db.SelectContext(ctx, &cards, getCards, req.Search, int32(req.PageSize), int32(offset))

	if err != nil {
		return nil, sharedErrors.ErrFailed("find all cards").WithInternal(err)
	}

	return cards, nil
}

func (r *cardQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllCards) ([]*db.GetActiveCardsWithCountRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetActiveCardsWithCountRow
	err := r.db.SelectContext(ctx, &res, getActiveCardsWithCount, req.Search, int32(req.PageSize), int32(offset))

	if err != nil {
		return nil, sharedErrors.ErrFailed("find active cards").WithInternal(err)
	}

	return res, nil
}

func (r *cardQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllCards) ([]*db.GetTrashedCardsWithCountRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetTrashedCardsWithCountRow
	err := r.db.SelectContext(ctx, &res, getTrashedCardsWithCount, req.Search, int32(req.PageSize), int32(offset))

	if err != nil {
		return nil, sharedErrors.ErrFailed("find trashed cards").WithInternal(err)
	}

	return res, nil
}

func (r *cardQueryRepository) FindById(ctx context.Context, card_id int) (*db.GetCardByIDRow, error) {
	var res db.GetCardByIDRow
	err := r.db.GetContext(ctx, &res, getCardByID, int32(card_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &res, nil
}

func (r *cardQueryRepository) FindCardByUserId(ctx context.Context, user_id int) (*db.GetCardByUserIDRow, error) {
	var res db.GetCardByUserIDRow
	err := r.db.GetContext(ctx, &res, getCardByUserID, int32(user_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &res, nil
}

func (r *cardQueryRepository) FindCardByCardNumber(ctx context.Context, card_number string) (*db.GetCardByCardNumberRow, error) {
	var res db.GetCardByCardNumberRow
	err := r.db.GetContext(ctx, &res, getCardByCardNumber, card_number)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &res, nil
}

func (r *cardQueryRepository) FindUserCardByCardNumber(ctx context.Context, card_number string) (*db.GetUserEmailByCardNumberRow, error) {
	var res db.GetUserEmailByCardNumberRow
	err := r.db.GetContext(ctx, &res, getUserEmailByCardNumber, card_number)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &res, nil
}
