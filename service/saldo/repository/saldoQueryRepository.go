package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/database/schema"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/jmoiron/sqlx"
)

// SQL source of truth (formerly database/query/saldo.sql).
const (
	getSaldos = `SELECT
    saldo_id,
    card_number,
    total_balance,
    withdraw_amount,
    withdraw_time,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM saldos
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR card_number ILIKE '%' || $1 || '%'
    )
ORDER BY saldo_id
LIMIT $2
OFFSET $3`

	getSaldoByID = `SELECT
    saldo_id,
    card_number,
    total_balance,
    withdraw_amount,
    withdraw_time,
    created_at,
    updated_at
FROM saldos
WHERE
    saldo_id = $1
    AND deleted_at IS NULL`

	getActiveSaldos = `SELECT
    saldo_id,
    card_number,
    total_balance,
    withdraw_amount,
    withdraw_time,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM saldos
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR card_number ILIKE '%' || $1 || '%'
    )
ORDER BY saldo_id
LIMIT $2
OFFSET $3`

	getTrashedSaldos = `SELECT
    saldo_id,
    card_number,
    total_balance,
    withdraw_amount,
    withdraw_time,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM saldos
WHERE
    deleted_at IS NOT NULL
    AND (
        $1::TEXT IS NULL
        OR card_number ILIKE '%' || $1 || '%'
    )
ORDER BY saldo_id
LIMIT $2
OFFSET $3`

	getSaldoByCardNumber = `SELECT * FROM saldos
WHERE card_number = $1 AND deleted_at IS NULL
ORDER BY saldo_id DESC
LIMIT 1`

	// Named queries kept for SQL source-of-truth parity (not yet wired to a
	// repository method):
	getTrashedSaldoByID = `SELECT * FROM saldos WHERE saldo_id = $1 AND deleted_at IS NOT NULL`

	getMonthlyTotalSaldoBalance = `WITH
    monthly_data AS (
        SELECT EXTRACT(
                YEAR
                FROM s.created_at
            )::integer AS year, EXTRACT(
                MONTH
                FROM s.created_at
            )::integer AS month, COALESCE(SUM(s.total_balance), 0) AS total_balance
        FROM saldos s
        WHERE
            s.deleted_at IS NULL
            AND (
                (
                    s.created_at >= $1::timestamp
                    AND s.created_at <= $2::timestamp
                )
                OR (
                    s.created_at >= $3::timestamp
                    AND s.created_at <= $4::timestamp
                )
            )
        GROUP BY
            EXTRACT(
                YEAR
                FROM s.created_at
            ),
            EXTRACT(
                MONTH
                FROM s.created_at
            )
    ),
    formatted_data AS (
        SELECT year::text, TO_CHAR(
                TO_DATE(month::text, 'MM'), 'Mon'
            ) AS month, total_balance::integer
        FROM monthly_data
        UNION ALL
        SELECT EXTRACT(
                YEAR
                FROM $1::timestamp
            )::text AS year, TO_CHAR($1::timestamp, 'Mon') AS month, 0::integer AS total_balance
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
        SELECT EXTRACT(
                YEAR
                FROM $3::timestamp
            )::text AS year, TO_CHAR($3::timestamp, 'Mon') AS month, 0::integer AS total_balance
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

	getYearlyTotalSaldoBalances = `WITH
    yearly_data AS (
        SELECT EXTRACT(
                YEAR
                FROM s.created_at
            )::integer AS year, COALESCE(SUM(s.total_balance), 0)::bigint AS total_balance
        FROM saldos s
        WHERE
            s.deleted_at IS NULL
            AND (
                EXTRACT(
                    YEAR
                    FROM s.created_at
                ) = $1::integer
                OR EXTRACT(
                    YEAR
                    FROM s.created_at
                ) = $1::integer - 1
            )
        GROUP BY
            EXTRACT(
                YEAR
                FROM s.created_at
            )
    ),
    formatted_data AS (
        SELECT year::text, total_balance::integer
        FROM yearly_data
        UNION ALL
        SELECT $1::text AS year, 0::integer AS total_balance
        WHERE
            NOT EXISTS (
                SELECT 1
                FROM yearly_data
                WHERE
                    year = $1::integer
            )
        UNION ALL
        SELECT ($1::integer - 1)::text AS year, 0::integer AS total_balance
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

	getMonthlySaldoBalances = `WITH
    months AS (
        SELECT generate_series(
                date_trunc('year', $1::timestamp), date_trunc('year', $1::timestamp) + interval '1 year' - interval '1 day', interval '1 month'
            ) AS month
    )
SELECT TO_CHAR(m.month, 'Mon') AS month, COALESCE(SUM(s.total_balance), 0)::bigint AS total_balance
FROM months m
    LEFT JOIN saldos s ON EXTRACT(
        MONTH
        FROM s.created_at
    ) = EXTRACT(
        MONTH
        FROM m.month
    )
    AND EXTRACT(
        YEAR
        FROM s.created_at
    ) = EXTRACT(
        YEAR
        FROM m.month
    )
    AND s.deleted_at IS NULL
GROUP BY
    m.month
ORDER BY m.month`

	getYearlySaldoBalances = `WITH
    last_five_years AS (
        SELECT EXTRACT(
                YEAR
                FROM s.created_at
            ) AS year, SUM(s.total_balance) AS total_balance
        FROM saldos s
        WHERE
            s.deleted_at IS NULL
            AND EXTRACT(
                YEAR
                FROM s.created_at
            ) >= $1 - 4
            AND EXTRACT(
                YEAR
                FROM s.created_at
            ) <= $1
        GROUP BY
            EXTRACT(
                YEAR
                FROM s.created_at
            )
    )
SELECT year, total_balance
FROM last_five_years
ORDER BY year`
)

type saldoQueryRepository struct {
	db *sqlx.DB
}

func NewSaldoQueryRepository(db *sqlx.DB) SaldoQueryRepository {
	return &saldoQueryRepository{
		db: db,
	}
}

func (r *saldoQueryRepository) ListReconciliationQueue(ctx context.Context, status string, limit int32) ([]*db.ReconciliationQueueRow, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	res, err := db.ListReconciliationQueue(ctx, r.db, status, limit)
	if err != nil {
		return nil, sharedErrors.ErrFailed("list reconciliation queue").WithInternal(err)
	}
	return res, nil
}

func (r *saldoQueryRepository) ListLedgerEntries(ctx context.Context, cardNumber string, limit int32) ([]*db.LedgerEntry, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	res, err := db.ListLedgerEntries(ctx, r.db, cardNumber, limit)
	if err != nil {
		return nil, sharedErrors.ErrFailed("list ledger entries").WithInternal(err)
	}
	return res, nil
}

func (r *saldoQueryRepository) FindAllSaldos(ctx context.Context, req *requests.FindAllSaldos) ([]*db.GetSaldosRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var saldos []*db.GetSaldosRow
	err := r.db.SelectContext(ctx, &saldos, getSaldos,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, sharedErrors.ErrFailed("find all saldo records").WithInternal(err)
	}

	return saldos, nil
}

func (r *saldoQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllSaldos) ([]*db.GetActiveSaldosRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetActiveSaldosRow
	err := r.db.SelectContext(ctx, &res, getActiveSaldos,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, sharedErrors.ErrFailed("find active saldo records").WithInternal(err)
	}

	return res, nil
}

func (r *saldoQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllSaldos) ([]*db.GetTrashedSaldosRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var saldos []*db.GetTrashedSaldosRow
	err := r.db.SelectContext(ctx, &saldos, getTrashedSaldos,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, sharedErrors.ErrFailed("find trashed saldo records").WithInternal(err)
	}

	return saldos, nil
}

func (r *saldoQueryRepository) FindByCardNumber(ctx context.Context, card_number string) (*db.Saldo, error) {
	var res db.Saldo
	err := r.db.GetContext(ctx, &res, getSaldoByCardNumber, card_number)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFoundResponse("saldo record").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &res, nil
}

func (r *saldoQueryRepository) FindById(ctx context.Context, saldo_id int) (*db.GetSaldoByIDRow, error) {
	var res db.GetSaldoByIDRow
	err := r.db.GetContext(ctx, &res, getSaldoByID, int32(saldo_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFoundResponse("saldo record").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &res, nil
}
