package db

import (
	"context"
	"time"
)

// StuckTransaction is a transaction row the recovery worker should re-drive.
// StuckTransfer is a transfer row the recovery worker should re-drive.
type StuckTransfer struct {
	TransferID     int32     `db:"transfer_id"`
	OperationID    string    `db:"operation_id"`
	TransferFrom   string    `db:"transfer_from"`
	TransferTo     string    `db:"transfer_to"`
	TransferAmount int32     `db:"transfer_amount"`
	Status         string    `db:"status"`
	UpdatedAt      time.Time `db:"updated_at"`
}

const guardTransferStatus = `
		UPDATE transfers
		SET status = $3,
		    failure_reason = CASE WHEN $4 = '' THEN failure_reason ELSE $4 END,
		    updated_at = current_timestamp
		WHERE transfer_id = $1 AND status = $2 AND deleted_at IS NULL
		RETURNING transfer_id, transfer_no, transfer_from, transfer_to,
		          transfer_amount, transfer_time, status, created_at, updated_at`

const listStuckTransfers = `
		SELECT transfer_id, operation_id::text, transfer_from, transfer_to, transfer_amount, status, updated_at
		FROM transfers
		WHERE status IN ('processing', 'compensating', 'unknown')
		  AND deleted_at IS NULL
		  AND updated_at < current_timestamp - ($1 * interval '1 second')
		ORDER BY updated_at
		LIMIT $2`

// guardQueryer is the executor surface needed by the operation-state queries.
// Both *sqlx.DB and *sqlx.Tx satisfy it.
type guardQueryer interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

// GuardTransferStatus is the transfer equivalent of GuardTransactionStatus.
func GuardTransferStatus(ctx context.Context, q guardQueryer, id int32, fromStatus, toStatus, reason string) (*UpdateTransferStatusRow, error) {
	var row UpdateTransferStatusRow
	err := q.GetContext(ctx, &row, guardTransferStatus, id, fromStatus, toStatus, reason)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// ListStuckTransfers returns transfers stuck in a recoverable state.
func ListStuckTransfers(ctx context.Context, q guardQueryer, olderThan time.Duration, maxRows int32) ([]*StuckTransfer, error) {
	var out []*StuckTransfer
	err := q.SelectContext(ctx, &out, listStuckTransfers, int64(olderThan.Seconds()), maxRows)
	return out, err
}
