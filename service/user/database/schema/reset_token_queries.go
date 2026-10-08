package db

const createResetToken = `INSERT INTO
    reset_tokens (user_id, token, expiry_date)
VALUES ($1, $2, $3)
RETURNING
    user_id,
    token,
    expiry_date`

const deleteResetToken = `DELETE FROM reset_tokens WHERE user_id = $1`

const getResetToken = `SELECT user_id, token, expiry_date
FROM reset_tokens
WHERE
    token = $1`
