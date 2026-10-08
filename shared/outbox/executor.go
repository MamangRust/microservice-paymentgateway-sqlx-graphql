package outbox

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// sqlxDB adapts *sqlx.DB to DBTX: the *sql.Row it returns is a concrete type,
// while the interface exposes the Scan-only Row surface.
type sqlxDB struct{ db *sqlx.DB }

func (w sqlxDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return w.db.ExecContext(ctx, query, args...)
}

func (w sqlxDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return w.db.QueryContext(ctx, query, args...)
}

func (w sqlxDB) QueryRowContext(ctx context.Context, query string, args ...any) Row {
	return w.db.QueryRowContext(ctx, query, args...)
}

// sqlxTx adapts *sqlx.Tx to DBTX (see sqlxDB).
type sqlxTx struct{ tx *sqlx.Tx }

func (w sqlxTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return w.tx.ExecContext(ctx, query, args...)
}

func (w sqlxTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return w.tx.QueryContext(ctx, query, args...)
}

func (w sqlxTx) QueryRowContext(ctx context.Context, query string, args ...any) Row {
	return w.tx.QueryRowContext(ctx, query, args...)
}

// FromDB wraps a *sqlx.DB handle as the outbox DBTX executor.
func FromDB(db *sqlx.DB) DBTX {
	if db == nil {
		return nil
	}
	return sqlxDB{db}
}

// FromTx wraps a *sqlx.Tx as the outbox DBTX executor (in-transaction insert).
func FromTx(tx *sqlx.Tx) DBTX {
	if tx == nil {
		return nil
	}
	return sqlxTx{tx}
}
