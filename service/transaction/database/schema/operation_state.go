package db

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

// StuckTransaction is a transaction row the recovery worker should re-drive.
type StuckTransaction struct {
	TransactionID int32     `db:"transaction_id"`
	OperationID   string    `db:"operation_id"`
	CardNumber    string    `db:"card_number"`
	Amount        int32     `db:"amount"`
	MerchantID    int32     `db:"merchant_id"`
	Status        string    `db:"status"`
	UpdatedAt     time.Time `db:"updated_at"`
}

const guardTransactionStatus = `
		UPDATE transactions
		SET status = $3,
		    failure_reason = CASE WHEN $4 = '' THEN failure_reason ELSE $4 END,
		    updated_at = current_timestamp
		WHERE transaction_id = $1 AND status = $2 AND deleted_at IS NULL
		RETURNING transaction_id, transaction_no, card_number, amount, payment_method,
		          merchant_id, transaction_time, status, created_at, updated_at`

// GuardTransactionStatus atomically moves a transaction from an expected status
// to a new one. It returns sql.ErrNoRows when the row is not in the expected
// status (another actor already moved it), which callers use to detect races.
func GuardTransactionStatus(ctx context.Context, database *sqlx.DB, id int32, fromStatus, toStatus, reason string) (*UpdateTransactionStatusRow, error) {
	var row UpdateTransactionStatusRow
	if err := database.GetContext(ctx, &row, guardTransactionStatus, id, fromStatus, toStatus, reason); err != nil {
		return nil, err
	}
	return &row, nil
}

const listStuckTransactions = `
		SELECT transaction_id, operation_id::text, card_number, amount, merchant_id, status, updated_at
		FROM transactions
		WHERE status IN ('processing', 'compensating', 'unknown')
		  AND deleted_at IS NULL
		  AND updated_at < current_timestamp - ($1 * interval '1 second')
		ORDER BY updated_at
		LIMIT $2`

// ListStuckTransactions returns transactions stuck in a recoverable state that
// have not been touched for at least olderThan, limited to maxRows.
func ListStuckTransactions(ctx context.Context, database *sqlx.DB, olderThan time.Duration, maxRows int32) ([]*StuckTransaction, error) {
	var out []*StuckTransaction
	if err := database.SelectContext(ctx, &out, listStuckTransactions, int64(olderThan.Seconds()), maxRows); err != nil {
		return nil, err
	}
	return out, nil
}
