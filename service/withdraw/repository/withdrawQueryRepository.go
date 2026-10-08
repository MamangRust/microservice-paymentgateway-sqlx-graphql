package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/jmoiron/sqlx"
)

// SQL source of truth (formerly database/query/withdraw.sql).
const (
	getWithdraws = `SELECT
    withdraw_id,
    withdraw_no,
    card_number,
    withdraw_amount,
    withdraw_time,
    status,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM withdraws
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR card_number ILIKE '%' || $1 || '%'
        OR withdraw_amount::TEXT ILIKE '%' || $1 || '%'
        OR withdraw_time::TEXT ILIKE '%' || $1 || '%'
        OR status ILIKE '%' || $1 || '%'
    )
ORDER BY withdraw_time DESC
LIMIT $2
OFFSET $3`

	getActiveWithdraws = `SELECT
    withdraw_id,
    withdraw_no,
    card_number,
    withdraw_amount,
    withdraw_time,
    status,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM withdraws
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR card_number ILIKE '%' || $1 || '%'
        OR withdraw_amount::TEXT ILIKE '%' || $1 || '%'
        OR withdraw_time::TEXT ILIKE '%' || $1 || '%'
        OR status ILIKE '%' || $1 || '%'
    )
ORDER BY withdraw_time DESC
LIMIT $2
OFFSET $3`

	getTrashedWithdraws = `SELECT
    withdraw_id,
    withdraw_no,
    card_number,
    withdraw_amount,
    withdraw_time,
    status,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM withdraws
WHERE
    deleted_at IS NOT NULL
    AND (
        $1::TEXT IS NULL
        OR card_number ILIKE '%' || $1 || '%'
        OR withdraw_amount::TEXT ILIKE '%' || $1 || '%'
        OR withdraw_time::TEXT ILIKE '%' || $1 || '%'
        OR status ILIKE '%' || $1 || '%'
    )
ORDER BY withdraw_time DESC
LIMIT $2
OFFSET $3`

	getWithdrawByID = `SELECT
    withdraw_id,
    withdraw_no,
    card_number,
    withdraw_amount,
    withdraw_time,
    status,
    created_at,
    updated_at
FROM withdraws
WHERE
    withdraw_id = $1
    AND deleted_at IS NULL`

	getWithdrawsByCardNumber = `SELECT
    withdraw_id,
    withdraw_no,
    card_number,
    withdraw_amount,
    withdraw_time,
    status,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM withdraws
WHERE
    deleted_at IS NULL
    AND card_number = $1
    AND (
        $2::TEXT IS NULL
        OR CAST(withdraw_amount AS TEXT) ILIKE '%' || $2 || '%'
        OR TO_CHAR(
            withdraw_time,
            'YYYY-MM-DD HH24:MI:SS'
        ) ILIKE '%' || $2 || '%'
        OR status ILIKE '%' || $2 || '%'
    )
ORDER BY withdraw_time DESC
LIMIT $3
OFFSET $4`

	getTodayWithdrawSumByCardNumber = `SELECT COALESCE(SUM(withdraw_amount), 0)::bigint AS total_withdraw_amount
FROM withdraws
WHERE
    card_number = $1
    AND status = 'success'
    AND deleted_at IS NULL
    AND withdraw_time >= date_trunc('day', now())`

	// Named queries kept for SQL source-of-truth parity (not yet wired to a
	// repository method):
	getTrashedWithdrawByID = `SELECT *
FROM withdraws
WHERE
    withdraw_id = $1
    AND deleted_at IS NOT NULL`

	findAllWithdrawsByCardNumber = `SELECT w.withdraw_id, w.card_number, w.withdraw_amount, w.withdraw_time, w.created_at, w.updated_at, w.deleted_at
FROM withdraws w
WHERE
    w.card_number = $1
    AND w.deleted_at IS NULL
ORDER BY w.withdraw_time DESC`

	getMonthWithdrawStatusSuccess = `WITH
    monthly_data AS (
        SELECT
            EXTRACT(
                YEAR
                FROM t.withdraw_time
            )::integer AS year,
            EXTRACT(
                MONTH
                FROM t.withdraw_time
            )::integer AS month,
            COUNT(*) AS total_success,
            COALESCE(SUM(t.withdraw_amount), 0)::bigint AS total_amount
        FROM withdraws t
        WHERE
            t.deleted_at IS NULL
            AND t.status = 'success'
            AND (
                (
                    t.withdraw_time >= $1::timestamp
                    AND t.withdraw_time <= $2::timestamp
                )
                OR (
                    t.withdraw_time >= $3::timestamp
                    AND t.withdraw_time <= $4::timestamp
                )
            )
        GROUP BY
            EXTRACT(
                YEAR
                FROM t.withdraw_time
            ),
            EXTRACT(
                MONTH
                FROM t.withdraw_time
            )
    ),
    formatted_data AS (
        SELECT
            year::text,
            TO_CHAR(
                TO_DATE(month::text, 'MM'),
                'Mon'
            ) AS month,
            total_success,
            total_amount
        FROM monthly_data
        UNION ALL
        SELECT
            EXTRACT(
                YEAR
                FROM $1::timestamp
            )::text AS year,
            TO_CHAR($1::timestamp, 'Mon') AS month,
            0 AS total_success,
            0 AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM monthly_data
                WHERE
                    year = EXTRACT(
                        YEAR
                        FROM $1::timestamp
                    )::integer
                    AND month = EXTRACT(
                        MONTH
                        FROM $1::timestamp
                    )::integer
            )
        UNION ALL
        SELECT
            EXTRACT(
                YEAR
                FROM $3::timestamp
            )::text AS year,
            TO_CHAR($3::timestamp, 'Mon') AS month,
            0 AS total_success,
            0 AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM monthly_data
                WHERE
                    year = EXTRACT(
                        YEAR
                        FROM $3::timestamp
                    )::integer
                    AND month = EXTRACT(
                        MONTH
                        FROM $3::timestamp
                    )::integer
            )
    )
SELECT *
FROM formatted_data
ORDER BY year DESC, TO_DATE(month, 'Mon') DESC`

	getYearlyWithdrawStatusSuccess = `WITH
    yearly_data AS (
        SELECT
            EXTRACT(
                YEAR
                FROM t.withdraw_time
            )::integer AS year,
            COUNT(*) AS total_success,
            COALESCE(SUM(t.withdraw_amount), 0)::bigint AS total_amount
        FROM withdraws t
        WHERE
            t.deleted_at IS NULL
            AND t.status = 'success'
            AND (
                EXTRACT(
                    YEAR
                    FROM t.withdraw_time
                ) = $1::integer
                OR EXTRACT(
                    YEAR
                    FROM t.withdraw_time
                ) = $1::integer - 1
            )
        GROUP BY
            EXTRACT(
                YEAR
                FROM t.withdraw_time
            )
    ),
    formatted_data AS (
        SELECT
            year::text,
            total_success::integer,
            total_amount::integer
        FROM yearly_data
        UNION ALL
        SELECT
            $1::text AS year,
            0::integer AS total_success,
            0::integer AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM yearly_data
                WHERE
                    year = $1::integer
            )
        UNION ALL
        SELECT ($1::integer - 1)::text AS year,
            0::integer AS total_success,
            0::integer AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM yearly_data
                WHERE
                    year = $1::integer - 1
            )
    )
SELECT *
FROM formatted_data
ORDER BY year DESC`

	getMonthWithdrawStatusFailed = `WITH
    monthly_data AS (
        SELECT
            EXTRACT(
                YEAR
                FROM t.withdraw_time
            )::integer AS year,
            EXTRACT(
                MONTH
                FROM t.withdraw_time
            )::integer AS month,
            COUNT(*) AS total_failed,
            COALESCE(SUM(t.withdraw_amount), 0)::bigint AS total_amount
        FROM withdraws t
        WHERE
            t.deleted_at IS NULL
            AND t.status = 'failed'
            AND (
                (
                    t.withdraw_time >= $1::timestamp
                    AND t.withdraw_time <= $2::timestamp
                )
                OR (
                    t.withdraw_time >= $3::timestamp
                    AND t.withdraw_time <= $4::timestamp
                )
            )
        GROUP BY
            EXTRACT(
                YEAR
                FROM t.withdraw_time
            ),
            EXTRACT(
                MONTH
                FROM t.withdraw_time
            )
    ),
    formatted_data AS (
        SELECT
            year::text,
            TO_CHAR(
                TO_DATE(month::text, 'MM'),
                'Mon'
            ) AS month,
            total_failed,
            total_amount
        FROM monthly_data
        UNION ALL
        SELECT
            EXTRACT(
                YEAR
                FROM $1::timestamp
            )::text AS year,
            TO_CHAR($1::timestamp, 'Mon') AS month,
            0 AS total_failed,
            0 AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM monthly_data
                WHERE
                    year = EXTRACT(
                        YEAR
                        FROM $1::timestamp
                    )::integer
                    AND month = EXTRACT(
                        MONTH
                        FROM $1::timestamp
                    )::integer
            )
        UNION ALL
        SELECT
            EXTRACT(
                YEAR
                FROM $3::timestamp
            )::text AS year,
            TO_CHAR($3::timestamp, 'Mon') AS month,
            0 AS total_failed,
            0 AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM monthly_data
                WHERE
                    year = EXTRACT(
                        YEAR
                        FROM $3::timestamp
                    )::integer
                    AND month = EXTRACT(
                        MONTH
                        FROM $3::timestamp
                    )::integer
            )
    )
SELECT *
FROM formatted_data
ORDER BY year DESC, TO_DATE(month, 'Mon') DESC`

	getYearlyWithdrawStatusFailed = `WITH
    yearly_data AS (
        SELECT
            EXTRACT(
                YEAR
                FROM t.withdraw_time
            )::integer AS year,
            COUNT(*) AS total_failed,
            COALESCE(SUM(t.withdraw_amount), 0)::bigint AS total_amount
        FROM withdraws t
        WHERE
            t.deleted_at IS NULL
            AND t.status = 'failed'
            AND (
                EXTRACT(
                    YEAR
                    FROM t.withdraw_time
                ) = $1::integer
                OR EXTRACT(
                    YEAR
                    FROM t.withdraw_time
                ) = $1::integer - 1
            )
        GROUP BY
            EXTRACT(
                YEAR
                FROM t.withdraw_time
            )
    ),
    formatted_data AS (
        SELECT
            year::text,
            total_failed::integer,
            total_amount::integer
        FROM yearly_data
        UNION ALL
        SELECT
            $1::text AS year,
            0::integer AS total_failed,
            0::integer AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM yearly_data
                WHERE
                    year = $1::integer
            )
        UNION ALL
        SELECT ($1::integer - 1)::text AS year,
            0::integer AS total_failed,
            0::integer AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM yearly_data
                WHERE
                    year = $1::integer - 1
            )
    )
SELECT *
FROM formatted_data
ORDER BY year DESC`

	getMonthWithdrawStatusSuccessCardNumber = `WITH
    monthly_data AS (
        SELECT
            EXTRACT(
                YEAR
                FROM t.withdraw_time
            )::integer AS year,
            EXTRACT(
                MONTH
                FROM t.withdraw_time
            )::integer AS month,
            COUNT(*) AS total_success,
            COALESCE(SUM(t.withdraw_amount), 0)::bigint AS total_amount
        FROM withdraws t
        WHERE
            t.deleted_at IS NULL
            AND t.status = 'success'
            AND t.card_number = $1
            AND (
                (
                    t.withdraw_time >= $2::timestamp
                    AND t.withdraw_time <= $3::timestamp
                )
                OR (
                    t.withdraw_time >= $4::timestamp
                    AND t.withdraw_time <= $5::timestamp
                )
            )
        GROUP BY
            EXTRACT(
                YEAR
                FROM t.withdraw_time
            ),
            EXTRACT(
                MONTH
                FROM t.withdraw_time
            )
    ),
    formatted_data AS (
        SELECT
            year::text,
            TO_CHAR(
                TO_DATE(month::text, 'MM'),
                'Mon'
            ) AS month,
            total_success,
            total_amount
        FROM monthly_data
        UNION ALL
        SELECT
            EXTRACT(
                YEAR
                FROM $2::timestamp
            )::text AS year,
            TO_CHAR($2::timestamp, 'Mon') AS month,
            0 AS total_success,
            0 AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM monthly_data
                WHERE
                    year = EXTRACT(
                        YEAR
                        FROM $2::timestamp
                    )::integer
                    AND month = EXTRACT(
                        MONTH
                        FROM $2::timestamp
                    )::integer
            )
        UNION ALL
        SELECT
            EXTRACT(
                YEAR
                FROM $3::timestamp
            )::text AS year,
            TO_CHAR($3::timestamp, 'Mon') AS month,
            0 AS total_success,
            0 AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM monthly_data
                WHERE
                    year = EXTRACT(
                        YEAR
                        FROM $3::timestamp
                    )::integer
                    AND month = EXTRACT(
                        MONTH
                        FROM $3::timestamp
                    )::integer
            )
    )
SELECT *
FROM formatted_data
ORDER BY year DESC, TO_DATE(month, 'Mon') DESC`

	getYearlyWithdrawStatusSuccessCardNumber = `WITH
    yearly_data AS (
        SELECT
            EXTRACT(
                YEAR
                FROM t.withdraw_time
            )::integer AS year,
            COUNT(*) AS total_success,
            COALESCE(SUM(t.withdraw_amount), 0)::bigint AS total_amount
        FROM withdraws t
        WHERE
            t.deleted_at IS NULL
            AND t.status = 'success'
            AND t.card_number = $1
            AND (
                EXTRACT(
                    YEAR
                    FROM t.withdraw_time
                ) = $2::integer
                OR EXTRACT(
                    YEAR
                    FROM t.withdraw_time
                ) = $2::integer - 1
            )
        GROUP BY
            EXTRACT(
                YEAR
                FROM t.withdraw_time
            )
    ),
    formatted_data AS (
        SELECT
            year::text,
            total_success::integer,
            total_amount::integer
        FROM yearly_data
        UNION ALL
        SELECT
            $2::text AS year,
            0::integer AS total_success,
            0::integer AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM yearly_data
                WHERE
                    year = $2::integer
            )
        UNION ALL
        SELECT ($2::integer - 1)::text AS year,
            0::integer AS total_success,
            0::integer AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM yearly_data
                WHERE
                    year = $2::integer - 1
            )
    )
SELECT *
FROM formatted_data
ORDER BY year DESC`

	getMonthWithdrawStatusFailedCardNumber = `WITH
    monthly_data AS (
        SELECT
            EXTRACT(
                YEAR
                FROM t.withdraw_time
            )::integer AS year,
            EXTRACT(
                MONTH
                FROM t.withdraw_time
            )::integer AS month,
            COUNT(*) AS total_failed,
            COALESCE(SUM(t.withdraw_amount), 0)::bigint AS total_amount
        FROM withdraws t
        WHERE
            t.deleted_at IS NULL
            AND t.status = 'failed'
            AND t.card_number = $1
            AND (
                (
                    t.withdraw_time >= $2::timestamp
                    AND t.withdraw_time <= $3::timestamp
                )
                OR (
                    t.withdraw_time >= $4::timestamp
                    AND t.withdraw_time <= $5::timestamp
                )
            )
        GROUP BY
            EXTRACT(
                YEAR
                FROM t.withdraw_time
            ),
            EXTRACT(
                MONTH
                FROM t.withdraw_time
            )
    ),
    formatted_data AS (
        SELECT
            year::text,
            TO_CHAR(
                TO_DATE(month::text, 'MM'),
                'Mon'
            ) AS month,
            total_failed,
            total_amount
        FROM monthly_data
        UNION ALL
        SELECT
            EXTRACT(
                YEAR
                FROM $2::timestamp
            )::text AS year,
            TO_CHAR($2::timestamp, 'Mon') AS month,
            0 AS total_failed,
            0 AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM monthly_data
                WHERE
                    year = EXTRACT(
                        YEAR
                        FROM $2::timestamp
                    )::integer
                    AND month = EXTRACT(
                        MONTH
                        FROM $2::timestamp
                    )::integer
            )
        UNION ALL
        SELECT
            EXTRACT(
                YEAR
                FROM $3::timestamp
            )::text AS year,
            TO_CHAR($3::timestamp, 'Mon') AS month,
            0 AS total_failed,
            0 AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM monthly_data
                WHERE
                    year = EXTRACT(
                        YEAR
                        FROM $3::timestamp
                    )::integer
                    AND month = EXTRACT(
                        MONTH
                        FROM $3::timestamp
                    )::integer
            )
    )
SELECT *
FROM formatted_data
ORDER BY year DESC, TO_DATE(month, 'Mon') DESC`

	getYearlyWithdrawStatusFailedCardNumber = `WITH
    yearly_data AS (
        SELECT
            EXTRACT(
                YEAR
                FROM t.withdraw_time
            )::integer AS year,
            COUNT(*) AS total_failed,
            COALESCE(SUM(t.withdraw_amount), 0)::bigint AS total_amount
        FROM withdraws t
        WHERE
            t.deleted_at IS NULL
            AND t.status = 'failed'
            AND t.card_number = $1
            AND (
                EXTRACT(
                    YEAR
                    FROM t.withdraw_time
                ) = $2::integer
                OR EXTRACT(
                    YEAR
                    FROM t.withdraw_time
                ) = $2::integer - 1
            )
        GROUP BY
            EXTRACT(
                YEAR
                FROM t.withdraw_time
            )
    ),
    formatted_data AS (
        SELECT
            year::text,
            total_failed::integer,
            total_amount::integer
        FROM yearly_data
        UNION ALL
        SELECT
            $2::text AS year,
            0::integer AS total_failed,
            0::integer AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM yearly_data
                WHERE
                    year = $2::integer
            )
        UNION ALL
        SELECT ($2::integer - 1)::text AS year,
            0::integer AS total_failed,
            0::integer AS total_amount
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM yearly_data
                WHERE
                    year = $2::integer - 1
            )
    )
SELECT *
FROM formatted_data
ORDER BY year DESC`

	getMonthlyWithdraws = `WITH
    months AS (
        SELECT generate_series(
                date_trunc('year', $1::timestamp), date_trunc('year', $1::timestamp) + interval '1 year' - interval '1 day', interval '1 month'
            ) AS month
    )
SELECT TO_CHAR(m.month, 'Mon') AS month, COALESCE(SUM(w.withdraw_amount), 0)::bigint AS total_withdraw_amount
FROM months m
    LEFT JOIN withdraws w ON EXTRACT(
        MONTH
        FROM w.withdraw_time
    ) = EXTRACT(
        MONTH
        FROM m.month
    )
    AND EXTRACT(
        YEAR
        FROM w.withdraw_time
    ) = EXTRACT(
        YEAR
        FROM m.month
    )
    AND w.deleted_at IS NULL
GROUP BY
    m.month
ORDER BY m.month`

	getYearlyWithdraws = `SELECT EXTRACT(
        YEAR
        FROM w.withdraw_time
    ) AS year, SUM(w.withdraw_amount) AS total_withdraw_amount
FROM withdraws w
WHERE
    w.deleted_at IS NULL
    AND EXTRACT(
        YEAR
        FROM w.withdraw_time
    ) >= $1 - 4
    AND EXTRACT(
        YEAR
        FROM w.withdraw_time
    ) <= $1
GROUP BY
    EXTRACT(
        YEAR
        FROM w.withdraw_time
    )
ORDER BY year`

	getMonthlyWithdrawsByCardNumber = `WITH
    months AS (
        SELECT generate_series(
                date_trunc('year', $2::timestamp), date_trunc('year', $2::timestamp) + interval '1 year' - interval '1 day', interval '1 month'
            ) AS month
    )
SELECT TO_CHAR(m.month, 'Mon') AS month, COALESCE(SUM(w.withdraw_amount), 0)::bigint AS total_withdraw_amount
FROM
    months m
    LEFT JOIN withdraws w ON EXTRACT(
        MONTH
        FROM w.withdraw_time
    ) = EXTRACT(
        MONTH
        FROM m.month
    )
    AND EXTRACT(
        YEAR
        FROM w.withdraw_time
    ) = EXTRACT(
        YEAR
        FROM m.month
    )
    AND w.card_number = $1
    AND w.deleted_at IS NULL
GROUP BY
    m.month
ORDER BY m.month`

	getYearlyWithdrawsByCardNumber = `SELECT EXTRACT(
        YEAR
        FROM w.created_at
    ) AS year, SUM(w.withdraw_amount) AS total_withdraw_amount
FROM withdraws w
WHERE
    w.deleted_at IS NULL
    AND w.card_number = $1
    AND EXTRACT(
        YEAR
        FROM w.created_at
    ) >= $2 - 4
    AND EXTRACT(
        YEAR
        FROM w.created_at
    ) <= $2
GROUP BY
    EXTRACT(
        YEAR
        FROM w.created_at
    )
ORDER BY year`
)

type withdrawQueryRepository struct {
	db *sqlx.DB
}

func NewWithdrawQueryRepository(db *sqlx.DB) WithdrawQueryRepository {
	return &withdrawQueryRepository{
		db: db,
	}
}

func (r *withdrawQueryRepository) FindAll(ctx context.Context, req *requests.FindAllWithdraws) ([]*db.GetWithdrawsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var withdraw []*db.GetWithdrawsRow
	err := r.db.SelectContext(ctx, &withdraw, getWithdraws,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, sharedErrors.ErrFailed("find all withdraws").WithInternal(err)
	}

	return withdraw, nil

}

func (r *withdrawQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllWithdraws) ([]*db.GetActiveWithdrawsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetActiveWithdrawsRow
	err := r.db.SelectContext(ctx, &res, getActiveWithdraws,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, sharedErrors.ErrFailed("find active withdraws").WithInternal(err)
	}

	return res, nil
}

func (r *withdrawQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllWithdraws) ([]*db.GetTrashedWithdrawsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetTrashedWithdrawsRow
	err := r.db.SelectContext(ctx, &res, getTrashedWithdraws,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, sharedErrors.ErrFailed("find trashed withdraws").WithInternal(err)
	}

	return res, nil
}

func (r *withdrawQueryRepository) FindAllByCardNumber(ctx context.Context, req *requests.FindAllWithdrawCardNumber) ([]*db.GetWithdrawsByCardNumberRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var withdraw []*db.GetWithdrawsByCardNumberRow
	err := r.db.SelectContext(ctx, &withdraw, getWithdrawsByCardNumber,
		req.CardNumber,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, sharedErrors.ErrFailed("find withdraws by card number").WithInternal(err)
	}

	return withdraw, nil

}

func (r *withdrawQueryRepository) FindById(ctx context.Context, id int) (*db.GetWithdrawByIDRow, error) {
	var withdraw db.GetWithdrawByIDRow
	err := r.db.GetContext(ctx, &withdraw, getWithdrawByID, int32(id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFoundResponse("withdraw").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &withdraw, nil
}

func (r *withdrawQueryRepository) GetTodayWithdrawSumByCardNumber(ctx context.Context, cardNumber string) (int64, error) {
	var total int64
	err := r.db.GetContext(ctx, &total, getTodayWithdrawSumByCardNumber, cardNumber)

	if err != nil {
		return 0, sharedErrors.ErrFailed("get today withdraw sum").WithInternal(err)
	}

	return total, nil
}
