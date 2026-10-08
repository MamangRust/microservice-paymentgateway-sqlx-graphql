package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jmoiron/sqlx"

	database "github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/randomvcc"
	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
)

const createCard = `INSERT INTO
    cards (
        user_id,
        card_number,
        card_type,
        expire_date,
        cvv,
        card_provider,
        created_at,
        updated_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        current_timestamp,
        current_timestamp
    )
RETURNING
    card_id,
    user_id,
    card_number,
    card_type,
    expire_date,
    cvv,
    card_provider,
    created_at,
    updated_at
`

const updateCard = `UPDATE cards
SET
    card_type = $2,
    expire_date = $3,
    cvv = $4,
    card_provider = $5,
    updated_at = current_timestamp
WHERE
    card_id = $1
    AND deleted_at IS NULL
RETURNING
    card_id,
    user_id,
    card_number,
    card_type,
    expire_date,
    cvv,
    card_provider,
    created_at,
    updated_at
`

const trashCard = `UPDATE cards
SET
    deleted_at = current_timestamp
WHERE
    card_id = $1
    AND deleted_at IS NULL
RETURNING
    card_id,
    user_id,
    card_number,
    card_type,
    expire_date,
    cvv,
    card_provider,
    created_at,
    updated_at,
    deleted_at
`

const restoreCard = `UPDATE cards
SET
    deleted_at = NULL
WHERE
    card_id = $1
    AND deleted_at IS NOT NULL
RETURNING
    card_id,
    user_id,
    card_number,
    card_type,
    expire_date,
    cvv,
    card_provider,
    created_at,
    updated_at,
    deleted_at
`

const deleteCardPermanently = `DELETE FROM cards WHERE card_id = $1 AND deleted_at IS NOT NULL
`

const restoreAllCards = `UPDATE cards
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL
`

const deleteAllPermanentCards = `DELETE FROM cards WHERE deleted_at IS NOT NULL
`

const toggleCardStatus = `UPDATE cards
SET status = CASE WHEN status = 'active' THEN 'suspended' ELSE 'active' END,
    updated_at = CURRENT_TIMESTAMP
WHERE card_id = $1 AND deleted_at IS NULL
RETURNING card_id, user_id, card_number, card_type, expire_date, cvv, card_provider,
          status, credit_limit, outstanding_balance, reward_points, created_at, updated_at
`

const updateCreditLimit = `UPDATE cards
SET credit_limit = $2, updated_at = CURRENT_TIMESTAMP
WHERE card_id = $1 AND deleted_at IS NULL
RETURNING card_id, user_id, card_number, card_type, expire_date, cvv, card_provider,
          status, credit_limit, outstanding_balance, reward_points, created_at, updated_at
`

const redeemRewardPoints = `UPDATE cards
SET reward_points = GREATEST(0, reward_points - $2), updated_at = CURRENT_TIMESTAMP
WHERE card_id = $1 AND deleted_at IS NULL AND reward_points >= $2
RETURNING card_id, user_id, card_number, card_type, expire_date, cvv, card_provider,
          status, credit_limit, outstanding_balance, reward_points, created_at, updated_at
`

type cardCommandRepository struct {
	db *sqlx.DB
}

func NewCardCommandRepository(db *sqlx.DB) CardCommandRepository {
	return &cardCommandRepository{db: db}
}

func (r *cardCommandRepository) CreateCard(ctx context.Context, request *requests.CreateCardRequest) (*db.CreateCardRow, error) {
	number, err := randomvcc.RandomCardNumber()
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	var res db.CreateCardRow
	err = r.db.GetContext(ctx, &res, createCard,
		int32(request.UserID),
		number,
		request.CardType,
		pgtype.Date{Time: request.ExpireDate, Valid: true},
		request.CVV,
		request.CardProvider,
	)
	if err != nil {
		if database.IsUniqueViolation(err) {
			return nil, sharedErrors.NewConflictError("card number already exists").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("create card").WithInternal(err)
	}
	return &res, nil
}

func (r *cardCommandRepository) UpdateCard(ctx context.Context, request *requests.UpdateCardRequest) (*db.UpdateCardRow, error) {
	var res db.UpdateCardRow
	err := r.db.GetContext(ctx, &res, updateCard,
		int32(request.CardID),
		request.CardType,
		pgtype.Date{Time: request.ExpireDate, Valid: true},
		request.CVV,
		request.CardProvider,
	)
	if err != nil {
		// The update query only returns rows for existing, non-deleted cards.
		// A missing row means the card does not exist (or was soft-deleted), so
		// surface a 404 instead of a misleading 500 INTERNAL from sql.ErrNoRows.
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("update card").WithInternal(err)
	}
	return &res, nil
}

func (r *cardCommandRepository) TrashedCard(ctx context.Context, cardID int) (*db.Card, error) {
	var res db.TrashCardRow
	err := r.db.GetContext(ctx, &res, trashCard, int32(cardID))
	if err != nil {
		// The trash query only returns rows for existing, non-deleted cards.
		// A missing row means the card does not exist, was soft-deleted, or is
		// already trashed — surface a 404 instead of a misleading 500 INTERNAL
		// from sql.ErrNoRows.
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("trash card").WithInternal(err)
	}
	return &db.Card{
		CardID: res.CardID, UserID: res.UserID, CardNumber: res.CardNumber,
		CardType: res.CardType, ExpireDate: res.ExpireDate, Cvv: res.Cvv,
		CardProvider: res.CardProvider, CreatedAt: res.CreatedAt,
		UpdatedAt: res.UpdatedAt, DeletedAt: res.DeletedAt,
	}, nil
}

func (r *cardCommandRepository) RestoreCard(ctx context.Context, cardID int) (*db.Card, error) {
	var res db.RestoreCardRow
	err := r.db.GetContext(ctx, &res, restoreCard, int32(cardID))
	if err != nil {
		// The restore query only returns rows for existing, currently-trashed
		// cards. A missing row means the card does not exist or is not in a
		// trashed state — surface a 404 instead of a misleading 500 INTERNAL
		// from sql.ErrNoRows.
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("restore card").WithInternal(err)
	}
	return &db.Card{
		CardID: res.CardID, UserID: res.UserID, CardNumber: res.CardNumber,
		CardType: res.CardType, ExpireDate: res.ExpireDate, Cvv: res.Cvv,
		CardProvider: res.CardProvider, CreatedAt: res.CreatedAt,
		UpdatedAt: res.UpdatedAt, DeletedAt: res.DeletedAt,
	}, nil
}

func (r *cardCommandRepository) DeleteCardPermanent(ctx context.Context, cardID int) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteCardPermanently, int32(cardID)); err != nil {
		// Map a missing row to 404 for consistency with the other card
		// mutations instead of a misleading 500 INTERNAL from sql.ErrNoRows.
		return false, sharedErrors.ErrNoRowsOrFailed(err, "card", "delete card permanently")
	}
	return true, nil
}

func (r *cardCommandRepository) RestoreAllCard(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, restoreAllCards); err != nil {
		return false, sharedErrors.ErrFailed("restore all cards").WithInternal(err)
	}
	return true, nil
}

func (r *cardCommandRepository) DeleteAllCardPermanent(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteAllPermanentCards); err != nil {
		return false, sharedErrors.ErrFailed("delete all cards permanently").WithInternal(err)
	}
	return true, nil
}

func (r *cardCommandRepository) ToggleCardStatus(ctx context.Context, request *requests.ToggleCardStatusRequest) (*db.UpdateCardStatusRow, error) {
	// The target schema extension performs the toggle atomically, so concurrent
	// requests cannot overwrite a status read by a previous request.
	var res db.ToggleCardStatusRow
	err := r.db.GetContext(ctx, &res, toggleCardStatus, int32(request.CardID))
	if err != nil {
		// The toggle query only returns rows for existing, non-deleted cards.
		// A missing row means the card does not exist (or was soft-deleted), so
		// surface a 404 instead of a misleading 500 INTERNAL from sql.ErrNoRows.
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("update card status").WithInternal(err)
	}
	return &db.UpdateCardStatusRow{
		CardID: res.CardID, UserID: res.UserID, CardNumber: res.CardNumber,
		CardType: res.CardType, ExpireDate: res.ExpireDate, Cvv: res.Cvv,
		CardProvider: res.CardProvider, Status: res.Status,
		CreditLimit: res.CreditLimit, OutstandingBalance: res.OutstandingBalance,
		RewardPoints: res.RewardPoints, CreatedAt: res.CreatedAt, UpdatedAt: res.UpdatedAt,
	}, nil
}

func (r *cardCommandRepository) UpdateCreditLimit(ctx context.Context, request *requests.UpdateCreditLimitRequest) (*db.UpdateCreditLimitRow, error) {
	var res db.UpdateCreditLimitRow
	err := r.db.GetContext(ctx, &res, updateCreditLimit,
		int32(request.CardID),
		int32(request.CreditLimit),
	)
	if err != nil {
		// The credit-limit query only returns rows for existing, non-deleted
		// cards. A missing row means the card does not exist (or was
		// soft-deleted), so surface a 404 instead of a misleading 500 INTERNAL
		// from sql.ErrNoRows.
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("update credit limit").WithInternal(err)
	}
	return &res, nil
}

func (r *cardCommandRepository) RedeemPoints(ctx context.Context, request *requests.RedeemPointsRequest) (*db.RedeemRewardPointsRow, error) {
	var res db.RedeemRewardPointsRow
	err := r.db.GetContext(ctx, &res, redeemRewardPoints,
		int32(request.CardID),
		int32(request.Points),
	)
	if err != nil {
		// The redeem query only returns rows when reward_points >= points. A
		// missing row means either the card does not exist (or is deleted) or
		// the balance is insufficient. Distinguish the two so callers get a
		// 404 for an unknown card and a 400 for an exhausted balance instead of
		// a misleading 500 INTERNAL from sql.ErrNoRows.
		if errors.Is(err, sql.ErrNoRows) {
			var card db.GetCardByIDRow
			if cardErr := r.db.GetContext(ctx, &card, getCardByID, int32(request.CardID)); cardErr != nil {
				if errors.Is(cardErr, sql.ErrNoRows) {
					return nil, sharedErrors.ErrNotFoundResponse("card").WithInternal(err)
				}
				return nil, sharedErrors.ErrFailed("redeem reward points").WithInternal(cardErr)
			}
			return nil, sharedErrors.NewBadRequestError("insufficient reward points").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("redeem reward points").WithInternal(err)
	}
	return &res, nil
}
