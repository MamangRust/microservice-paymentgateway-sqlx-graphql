package repository

// SQL statements migrated verbatim from database/query/transfer.sql and
// database/query/role.sql (sqlc source of truth), plus handwritten
// operation-state and idempotency queries previously in database/schema.

const createTransfer = `
INSERT INTO
    transfers (
        transfer_from,
        transfer_to,
        transfer_amount,
        transfer_time,
        status,
        created_at,
        updated_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        current_timestamp,
        current_timestamp
    )
RETURNING
    transfer_id,
    transfer_no,
    transfer_from,
    transfer_to,
    transfer_amount,
    transfer_time,
    status,
    created_at,
    updated_at
`

const deleteAllPermanentTransfers = `
DELETE FROM transfers WHERE deleted_at IS NOT NULL
`

const deleteTransferPermanently = `
DELETE FROM transfers
WHERE
    transfer_id = $1
    AND deleted_at IS NOT NULL
`

const getTransferByID = `
SELECT
    transfer_id,
    transfer_no,
    transfer_from,
    transfer_to,
    transfer_amount,
    transfer_time,
    status,
    created_at,
    updated_at
FROM transfers
WHERE
    transfer_id = $1
    AND deleted_at IS NULL
`

const getTransfers = `
SELECT
    transfer_id,
    transfer_no,
    transfer_from,
    transfer_to,
    transfer_amount,
    transfer_time,
    status,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM transfers
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR transfer_from ILIKE '%' || $1 || '%'
        OR transfer_to ILIKE '%' || $1 || '%'
    )
ORDER BY transfer_time DESC
LIMIT $2
OFFSET
    $3
`

const getActiveTransfers = `
SELECT
    transfer_id,
    transfer_no,
    transfer_from,
    transfer_to,
    transfer_amount,
    transfer_time,
    status,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM transfers
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR transfer_from ILIKE '%' || $1 || '%'
        OR transfer_to ILIKE '%' || $1 || '%'
    )
ORDER BY transfer_time DESC
LIMIT $2
OFFSET
    $3
`

const getTrashedTransfers = `
SELECT
    transfer_id,
    transfer_no,
    transfer_from,
    transfer_to,
    transfer_amount,
    transfer_time,
    status,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM transfers
WHERE
    deleted_at IS NOT NULL
    AND (
        $1::TEXT IS NULL
        OR transfer_from ILIKE '%' || $1 || '%'
        OR transfer_to ILIKE '%' || $1 || '%'
    )
ORDER BY transfer_time DESC
LIMIT $2
OFFSET
    $3
`

const getTransfersBySourceCard = `
SELECT
    transfer_id,
    transfer_no,
    transfer_from,
    transfer_to,
    transfer_amount,
    transfer_time,
    status,
    created_at,
    updated_at
FROM transfers
WHERE
    deleted_at IS NULL
    AND transfer_from = $1
ORDER BY transfer_time DESC
`

const getTransfersByDestinationCard = `
SELECT
    transfer_id,
    transfer_no,
    transfer_from,
    transfer_to,
    transfer_amount,
    transfer_time,
    status,
    created_at,
    updated_at
FROM transfers
WHERE
    deleted_at IS NULL
    AND transfer_to = $1
ORDER BY transfer_time DESC
`

const restoreAllTransfers = `
UPDATE transfers
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL
`

const restoreTransfer = `
UPDATE transfers
SET
    deleted_at = NULL
WHERE
    transfer_id = $1
    AND deleted_at IS NOT NULL
RETURNING transfer_id, transfer_no, transfer_from, transfer_to, transfer_amount, transfer_time, status, created_at, updated_at, deleted_at, operation_id, failure_reason
`

const trashTransfer = `
UPDATE transfers
SET
    deleted_at = current_timestamp
WHERE
    transfer_id = $1
    AND deleted_at IS NULL
RETURNING transfer_id, transfer_no, transfer_from, transfer_to, transfer_amount, transfer_time, status, created_at, updated_at, deleted_at, operation_id, failure_reason
`

const updateTransfer = `
UPDATE transfers
SET
    transfer_from = $2,
    transfer_to = $3,
    transfer_amount = $4,
    transfer_time = $5,
    updated_at = current_timestamp
WHERE
    transfer_id = $1
    AND deleted_at IS NULL
RETURNING
    transfer_id,
    transfer_no,
    transfer_from,
    transfer_to,
    transfer_amount,
    transfer_time,
    status,
    created_at,
    updated_at
`

const updateTransferAmount = `
UPDATE transfers
SET
    transfer_amount = $2,
    transfer_time = current_timestamp,
    updated_at = current_timestamp
WHERE
    transfer_id = $1
    AND deleted_at IS NULL
RETURNING
    transfer_id,
    transfer_no,
    transfer_from,
    transfer_to,
    transfer_amount,
    transfer_time,
    status,
    created_at,
    updated_at
`

const updateTransferStatus = `
UPDATE transfers
SET
    status = $2,
    updated_at = current_timestamp
WHERE
    transfer_id = $1
    AND deleted_at IS NULL
RETURNING
    transfer_id,
    transfer_no,
    transfer_from,
    transfer_to,
    transfer_amount,
    transfer_time,
    status,
    created_at,
    updated_at
`

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

const claimIdempotency = `
		INSERT INTO idempotency_records (scope, idempotency_key, request_hash, status)
		VALUES ($1, $2, $3, 'processing')
		ON CONFLICT (scope, idempotency_key) DO UPDATE
			SET request_hash = EXCLUDED.request_hash,
			    status = 'processing',
			    response_payload = NULL,
			    resource_id = NULL,
			    updated_at = current_timestamp,
			    expires_at = current_timestamp + INTERVAL '24 hours'
			WHERE idempotency_records.expires_at <= current_timestamp
		RETURNING ` + idempotencyColumns

const getIdempotency = `SELECT ` + idempotencyColumns + ` FROM idempotency_records WHERE scope=$1 AND idempotency_key=$2 AND expires_at > current_timestamp`

const completeIdempotency = `
		UPDATE idempotency_records
		SET status=$4, response_payload=$5, resource_id=$6, updated_at=current_timestamp
		WHERE scope=$1 AND idempotency_key=$2 AND request_hash=$3`

const releaseIdempotency = `DELETE FROM idempotency_records WHERE scope=$1 AND idempotency_key=$2 AND request_hash=$3 AND status='processing'`

const idempotencyColumns = `id, scope, idempotency_key, request_hash, operation_id, status, response_payload, resource_id, created_at, updated_at, expires_at`
