package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
)

const insertCardReward = `INSERT INTO card_rewards (
    card_number, txn_id, amount, mcc, points_earned, expires_at
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING reward_id, card_number, txn_id, amount, mcc, points_earned,
          expires_at, redeemed, created_at
`

const getCardRewardBalance = `SELECT COALESCE(SUM(points_earned), 0)::BIGINT AS balance
FROM card_rewards
WHERE card_number = $1 AND redeemed = FALSE AND expires_at > CURRENT_TIMESTAMP
`

const getCardRewardHistory = `SELECT reward_id, card_number, txn_id, amount, mcc, points_earned,
       expires_at, redeemed, created_at
FROM card_rewards
WHERE card_number = $1
ORDER BY created_at DESC
`

const getRedeemableRewardIds = `SELECT reward_id, points_earned
FROM card_rewards
WHERE card_number = $1 AND redeemed = FALSE AND expires_at > CURRENT_TIMESTAMP
ORDER BY created_at
`

const markRewardsRedeemed = `WITH updated AS (
    UPDATE card_rewards
    SET redeemed = TRUE
    WHERE reward_id = ANY($1::INT[])
    RETURNING points_earned
)
SELECT COALESCE(SUM(points_earned), 0)::BIGINT AS points
FROM updated
`

type cardRewardRepository struct {
	db *sqlx.DB
}

func NewCardRewardRepository(q *sqlx.DB) CardRewardRepository {
	return &cardRewardRepository{db: q}
}

func earnPoints(amount int64, mcc string) int32 {
	points := int32(amount / 10000)
	switch mcc {
	case "5812", "4511", "5541":
		points *= 2
	}
	if points < 1 {
		points = 1
	}
	return points
}

func (r *cardRewardRepository) EarnRewards(ctx context.Context, req *requests.EarnRewardsRequest) (*db.CardReward, error) {
	var res db.CardReward
	err := r.db.GetContext(ctx, &res, insertCardReward,
		req.CardNumber,
		req.TxnID,
		req.Amount,
		req.Mcc,
		earnPoints(req.Amount, req.Mcc),
		time.Now().AddDate(1, 0, 0),
	)
	if err != nil {
		return nil, sharedErrors.ErrFailed("earn card rewards").WithInternal(err)
	}
	return &res, nil
}

func (r *cardRewardRepository) GetBalance(ctx context.Context, cardNumber string) (int64, error) {
	var balance int64
	err := r.db.GetContext(ctx, &balance, getCardRewardBalance, cardNumber)
	if err != nil {
		return 0, sharedErrors.ErrFailed("get reward balance").WithInternal(err)
	}
	return balance, nil
}

func (r *cardRewardRepository) GetHistory(ctx context.Context, cardNumber string) ([]*db.CardReward, error) {
	var res []*db.CardReward
	err := r.db.SelectContext(ctx, &res, getCardRewardHistory, cardNumber)
	if err != nil {
		return nil, sharedErrors.ErrFailed("get reward history").WithInternal(err)
	}
	return res, nil
}

func (r *cardRewardRepository) RedeemRewards(ctx context.Context, cardNumber string, points int64) (int64, error) {
	var rows []*db.GetRedeemableRewardIdsRow
	err := r.db.SelectContext(ctx, &rows, getRedeemableRewardIds, cardNumber)
	if err != nil {
		return 0, sharedErrors.ErrFailed("find redeemable rewards").WithInternal(err)
	}

	var ids []int32
	var total int64
	for _, row := range rows {
		if total >= points {
			break
		}
		ids = append(ids, row.RewardID)
		total += int64(row.PointsEarned)
	}
	if total < points {
		return 0, sharedErrors.ErrBadRequest.WithMessage("Insufficient reward points")
	}

	var redeemed int64
	err = r.db.GetContext(ctx, &redeemed, markRewardsRedeemed, ids)
	if err != nil {
		return 0, sharedErrors.ErrFailed("redeem rewards").WithInternal(err)
	}
	return redeemed, nil
}
