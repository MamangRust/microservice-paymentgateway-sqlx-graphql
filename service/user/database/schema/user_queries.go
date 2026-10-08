package db

const createUser = `INSERT INTO
    users (
        firstname,
        lastname,
        email,
        password,
        verification_code,
        is_verified,
        created_at,
        updated_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        current_timestamp,
        current_timestamp
    )
RETURNING
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at`

const deleteAllPermanentUsers = `DELETE FROM users WHERE deleted_at IS NOT NULL`

const deleteUserPermanently = `DELETE FROM users WHERE user_id = $1 AND deleted_at IS NOT NULL`

const getActiveUsersWithPagination = `SELECT
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM users
WHERE
    deleted_at IS NOT NULL
    AND (
        $1::TEXT IS NULL
        OR firstname ILIKE '%' || $1 || '%'
        OR lastname ILIKE '%' || $1 || '%'
        OR email ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET $3`

const getTrashedUserByID = `SELECT user_id, firstname, lastname, email, password, verification_code, is_verified, created_at, updated_at, deleted_at
FROM users
WHERE
    user_id = $1
    AND deleted_at IS NOT NULL`

const getTrashedUsersWithPagination = `SELECT
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM users
WHERE
    deleted_at IS NOT NULL
    AND (
        $1::TEXT IS NULL
        OR firstname ILIKE '%' || $1 || '%'
        OR lastname ILIKE '%' || $1 || '%'
        OR email ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET $3`

const getUserByEmail = `SELECT user_id, firstname, lastname, email, password, created_at, updated_at
FROM users
WHERE
    email = $1
    AND deleted_at IS NULL`

const getUserByEmailAndVerified = `SELECT
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at
FROM users
WHERE
    email = $1
    AND is_verified = true
    AND deleted_at IS NULL`

const getUserByEmailWithPassword = `SELECT user_id, email, password
FROM users
WHERE
    email = $1
    AND deleted_at IS NULL`

const getUserByID = `SELECT
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at
FROM users
WHERE
    user_id = $1
    AND deleted_at IS NULL`

const getUserByVerificationCode = `SELECT
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at
FROM users
WHERE
    verification_code = $1`

const getUsersWithPagination = `SELECT
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM users
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR firstname ILIKE '%' || $1 || '%'
        OR lastname ILIKE '%' || $1 || '%'
        OR email ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET $3`

const restoreAllUsers = `UPDATE users
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL`

const restoreUser = `UPDATE users
SET
    deleted_at = NULL,
    updated_at = current_timestamp
WHERE
    user_id = $1
    AND deleted_at IS NOT NULL
RETURNING
    user_id,
    firstname,
    lastname,
    email,
    created_at,
    updated_at,
    deleted_at`

const searchUsersByEmail = `SELECT user_id, firstname, lastname, email, password, verification_code, is_verified, created_at, updated_at, deleted_at
FROM users
WHERE
    deleted_at IS NULL
    AND email ILIKE '%' || $1 || '%'
ORDER BY created_at DESC`

const trashUser = `UPDATE users
SET
    deleted_at = current_timestamp,
    updated_at = current_timestamp
WHERE
    user_id = $1
    AND deleted_at IS NULL
RETURNING
    user_id,
    firstname,
    lastname,
    email,
    created_at,
    updated_at,
    deleted_at`

const updateUser = `UPDATE users
SET
    firstname = $2,
    lastname = $3,
    email = $4,
    password = $5,
    updated_at = current_timestamp
WHERE
    user_id = $1
    AND deleted_at IS NULL
RETURNING
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at`

const updateUserIsVerified = `UPDATE users
SET
    is_verified = $2,
    updated_at = current_timestamp
WHERE
    user_id = $1
    AND deleted_at IS NULL
RETURNING
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at`

const updateUserPassword = `UPDATE users
SET
    password = $2,
    updated_at = current_timestamp
WHERE
    user_id = $1
    AND deleted_at IS NULL
RETURNING
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at`
