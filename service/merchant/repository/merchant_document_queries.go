package repository

// Queries from database/query/merchant_document.sql (verbatim, minus -- name lines).

// Kept as the SQL source of truth after the sqlc -> sqlx migration.

// CreateMerchantDocument :one

const createMerchantDocument = `INSERT INTO
    merchant_documents (
        merchant_id,
        document_type,
        document_url,
        status,
        note,
        uploaded_at,
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
    document_id,
    merchant_id,
    document_type,
    document_url,
    status,
    note,
    uploaded_at,
    created_at,
    updated_at`

// DeleteAllPermanentMerchantDocuments :exec

const deleteAllPermanentMerchantDocuments = `DELETE FROM merchant_documents WHERE deleted_at IS NOT NULL`

// DeleteMerchantDocumentPermanently :exec

const deleteMerchantDocumentPermanently = `DELETE FROM merchant_documents
WHERE
    document_id = $1
    AND deleted_at IS NOT NULL`

// GetActiveMerchantDocuments :many

const getActiveMerchantDocuments = `SELECT
    document_id,
    merchant_id,
    document_type,
    document_url,
    status,
    note,
    uploaded_at,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM merchant_documents
WHERE
    deleted_at IS NULL
    AND status != 'deleted'
    AND (
        $1::TEXT IS NULL
        OR document_type ILIKE '%' || $1 || '%'
        OR status ILIKE '%' || $1 || '%'
        OR note ILIKE '%' || $1 || '%'
    )
ORDER BY document_id
LIMIT $2
OFFSET
    $3`

// GetMerchantDocument :one

const getMerchantDocument = `SELECT
    document_id,
    merchant_id,
    document_type,
    document_url,
    status,
    note,
    uploaded_at,
    created_at,
    updated_at
FROM merchant_documents
WHERE
    document_id = $1
    AND deleted_at IS NULL`

// GetMerchantDocuments :many

const getMerchantDocuments = `SELECT
    document_id,
    merchant_id,
    document_type,
    document_url,
    status,
    note,
    uploaded_at,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM merchant_documents
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR document_type ILIKE '%' || $1 || '%'
        OR status ILIKE '%' || $1 || '%'
        OR note ILIKE '%' || $1 || '%'
    )
ORDER BY document_id
LIMIT $2
OFFSET
    $3`

// GetTrashedMerchantDocuments :many

const getTrashedMerchantDocuments = `SELECT
    document_id,
    merchant_id,
    document_type,
    document_url,
    status,
    note,
    uploaded_at,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM merchant_documents
WHERE
    deleted_at IS NOT NULL
    AND (
        $1::TEXT IS NULL
        OR document_type ILIKE '%' || $1 || '%'
        OR status ILIKE '%' || $1 || '%'
        OR note ILIKE '%' || $1 || '%'
    )
ORDER BY document_id
LIMIT $2
OFFSET
    $3`

// RestoreAllMerchantDocuments :exec

const restoreAllMerchantDocuments = `UPDATE merchant_documents
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL`

// RestoreMerchantDocument :one

const restoreMerchantDocument = `UPDATE merchant_documents
SET
    deleted_at = NULL
WHERE
    document_id = $1
    AND deleted_at IS NOT NULL
RETURNING
    document_id,
    merchant_id,
    document_type,
    document_url,
    status,
    note,
    uploaded_at,
    created_at,
    updated_at,
    deleted_at`

// TrashMerchantDocument :one

const trashMerchantDocument = `UPDATE merchant_documents
SET
    deleted_at = current_timestamp
WHERE
    document_id = $1
    AND deleted_at IS NULL
RETURNING
    document_id,
    merchant_id,
    document_type,
    document_url,
    status,
    note,
    uploaded_at,
    created_at,
    updated_at,
    deleted_at`

// UpdateMerchantDocument :one

const updateMerchantDocument = `UPDATE merchant_documents
SET
    document_type = $2,
    document_url = $3,
    status = $4,
    note = $5,
    updated_at = current_timestamp
WHERE
    document_id = $1
    AND deleted_at IS NULL
RETURNING
    document_id,
    merchant_id,
    document_type,
    document_url,
    status,
    note,
    uploaded_at,
    created_at,
    updated_at`

// UpdateMerchantDocumentStatus :one

const updateMerchantDocumentStatus = `UPDATE merchant_documents
SET
    status = $2,
    note = $3,
    updated_at = current_timestamp
WHERE
    document_id = $1
    AND deleted_at IS NULL
RETURNING
    document_id,
    merchant_id,
    document_type,
    document_url,
    status,
    note,
    uploaded_at,
    created_at,
    updated_at`
