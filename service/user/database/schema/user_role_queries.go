package db

const assignRoleToUser = `INSERT INTO user_roles (
    user_id,
    role_id,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    current_timestamp,
    current_timestamp
) RETURNING
    user_role_id,
    user_id,
    role_id,
    created_at,
    updated_at,
    deleted_at`

const getTrashedUserRoles = `SELECT
    ur.user_role_id,
    ur.user_id,
    ur.role_id,
    r.role_name,
    ur.created_at,
    ur.updated_at,
    ur.deleted_at
FROM
    user_roles ur
JOIN
    roles r ON ur.role_id = r.role_id
WHERE
    ur.user_id = $1
    AND ur.deleted_at IS NOT NULL
ORDER BY
    ur.deleted_at DESC`

const removeRoleFromUser = `DELETE FROM user_roles
WHERE
    user_id = $1
    AND role_id = $2`

const restoreUserRole = `UPDATE user_roles
SET
    deleted_at = NULL
WHERE
    user_role_id = $1`

const trashUserRole = `UPDATE user_roles
SET
    deleted_at = current_timestamp
WHERE
    user_role_id = $1`
