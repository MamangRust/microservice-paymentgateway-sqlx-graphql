package repository

// Queries from database/query/merchant.sql (verbatim, minus -- name lines).

// Kept as the SQL source of truth after the sqlc -> sqlx migration.

// CreateMerchant :one

const createMerchant = `INSERT INTO
    merchants (
        name,
        api_key,
        user_id,
        status,
        created_at,
        updated_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        current_timestamp,
        current_timestamp
    )
RETURNING
    merchant_id,
    merchant_no,
    name,
    api_key,
    user_id,
    status,
    created_at,
    updated_at`

// DeleteAllPermanentMerchants :exec

const deleteAllPermanentMerchants = `DELETE FROM merchants WHERE deleted_at IS NOT NULL`

// DeleteMerchantPermanently :exec

const deleteMerchantPermanently = `DELETE FROM merchants
WHERE
    merchant_id = $1
    AND deleted_at IS NOT NULL`

// GetActiveMerchants :many

const getActiveMerchants = `SELECT
    merchant_id,
    merchant_no,
    name,
    api_key,
    user_id,
    status,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM merchants
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR name ILIKE '%' || $1 || '%'
        OR api_key ILIKE '%' || $1 || '%'
        OR status ILIKE '%' || $1 || '%'
    )
ORDER BY merchant_id
LIMIT $2
OFFSET
    $3`

// GetMerchantByApiKey :one

const getMerchantByApiKey = `SELECT
    merchant_id,
    merchant_no,
    name,
    api_key,
    user_id,
    status,
    created_at,
    updated_at
FROM merchants
WHERE
    api_key = $1
    AND deleted_at IS NULL`

// GetMerchantByID :one

const getMerchantByID = `SELECT
    merchant_id,
    merchant_no,
    name,
    api_key,
    user_id,
    status,
    created_at,
    updated_at
FROM merchants
WHERE
    merchant_id = $1
    AND deleted_at IS NULL`

// GetMerchantByName :one

const getMerchantByName = `SELECT
    merchant_id,
    merchant_no,
    name,
    api_key,
    user_id,
    status,
    created_at,
    updated_at
FROM merchants
WHERE
    name = $1
    AND deleted_at IS NULL`

// GetMerchants :many

const getMerchants = `SELECT
    merchant_id,
    merchant_no,
    name,
    api_key,
    user_id,
    status,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM merchants
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR name ILIKE '%' || $1 || '%'
        OR api_key ILIKE '%' || $1 || '%'
        OR status ILIKE '%' || $1 || '%'
    )
ORDER BY merchant_id
LIMIT $2
OFFSET
    $3`

// GetMerchantsByUserID :many

const getMerchantsByUserID = `SELECT
    merchant_id,
    merchant_no,
    name,
    api_key,
    user_id,
    status,
    created_at,
    updated_at
FROM merchants
WHERE
    user_id = $1
    AND deleted_at IS NULL`

// GetTrashedMerchantByID :one

const getTrashedMerchantByID = `SELECT merchant_id, merchant_no, name, api_key, user_id, status, created_at, updated_at, deleted_at
FROM merchants
WHERE
    merchant_id = $1
    AND deleted_at IS NOT NULL`

// GetTrashedMerchants :many

const getTrashedMerchants = `SELECT
    merchant_id,
    merchant_no,
    name,
    api_key,
    user_id,
    status,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM merchants
WHERE
    deleted_at IS NOT NULL
    AND (
        $1::TEXT IS NULL
        OR name ILIKE '%' || $1 || '%'
        OR api_key ILIKE '%' || $1 || '%'
        OR status ILIKE '%' || $1 || '%'
    )
ORDER BY merchant_id
LIMIT $2
OFFSET
    $3`

// RestoreAllMerchants :exec

const restoreAllMerchants = `UPDATE merchants
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL`

// RestoreMerchant :one

const restoreMerchant = `UPDATE merchants
SET
    deleted_at = NULL
WHERE
    merchant_id = $1
    AND deleted_at IS NOT NULL
RETURNING
    merchant_id,
    merchant_no,
    name,
    api_key,
    user_id,
    status,
    created_at,
    updated_at,
    deleted_at`

// TrashMerchant :one

const trashMerchant = `UPDATE merchants
SET
    deleted_at = current_timestamp
WHERE
    merchant_id = $1
    AND deleted_at IS NULL
RETURNING
    merchant_id,
    merchant_no,
    name,
    api_key,
    user_id,
    status,
    created_at,
    updated_at,
    deleted_at`

// UpdateMerchant :one

const updateMerchant = `UPDATE merchants
SET
    name = $2,
    user_id = $3,
    status = $4,
    updated_at = current_timestamp
WHERE
    merchant_id = $1
    AND deleted_at IS NULL
RETURNING
    merchant_id,
    merchant_no,
    name,
    api_key,
    user_id,
    status,
    created_at,
    updated_at`

// UpdateMerchantStatus :one

const updateMerchantStatus = `UPDATE merchants
SET
    status = $2,
    updated_at = current_timestamp
WHERE
    merchant_id = $1
    AND deleted_at IS NULL
RETURNING
    merchant_id,
    merchant_no,
    name,
    api_key,
    user_id,
    status,
    created_at,
    updated_at`
