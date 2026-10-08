package db

const createRole = `INSERT INTO
    roles (
        role_name,
        created_at,
        updated_at
    )
VALUES (
        $1,
        current_timestamp,
        current_timestamp
    )
RETURNING
    role_id,
    role_name,
    created_at,
    updated_at,
    deleted_at`

const deleteAllPermanentRoles = `DELETE FROM roles WHERE deleted_at IS NOT NULL`

const deletePermanentRole = `DELETE FROM roles
WHERE
    role_id = $1
    AND deleted_at IS NOT NULL`

const getActiveRoles = `SELECT
    role_id,
    role_name,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM roles
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR role_name ILIKE '%' || $1 || '%'
    )
ORDER BY created_at ASC
LIMIT $2
OFFSET $3`

const getRole = `SELECT
    role_id,
    role_name,
    created_at,
    updated_at,
    deleted_at
FROM roles
WHERE
    role_id = $1`

const getRoleByName = `SELECT
    role_id,
    role_name,
    created_at,
    updated_at,
    deleted_at
FROM roles
WHERE
    role_name = $1`

const getRoles = `SELECT
    role_id,
    role_name,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM roles
WHERE
    $1::TEXT IS NULL
    OR role_name ILIKE '%' || $1 || '%'
ORDER BY created_at ASC
LIMIT $2
OFFSET $3`

const getTrashedRoles = `SELECT
    role_id,
    role_name,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM roles
WHERE
    deleted_at IS NOT NULL
    AND (
        $1::TEXT IS NULL
        OR role_name ILIKE '%' || $1 || '%'
    )
ORDER BY deleted_at DESC
LIMIT $2
OFFSET $3`

const getUserRoles = `SELECT r.role_id, r.role_name, r.created_at, r.updated_at, r.deleted_at
FROM roles r
    JOIN user_roles ur ON ur.role_id = r.role_id
WHERE
    ur.user_id = $1
ORDER BY r.created_at ASC`

const restoreAllRoles = `UPDATE roles
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL`

const restoreRole = `UPDATE roles
SET
    deleted_at = NULL
WHERE
    role_id = $1
RETURNING
    role_id, role_name, created_at, updated_at, deleted_at`

const trashRole = `UPDATE roles
SET
    deleted_at = current_timestamp
WHERE
    role_id = $1
RETURNING
    role_id, role_name, created_at, updated_at, deleted_at`

const updateRole = `UPDATE roles
SET
    role_name = $2,
    updated_at = current_timestamp
WHERE
    role_id = $1
RETURNING
    role_id,
    role_name,
    created_at,
    updated_at,
    deleted_at`
