-- name: CreateUser :exec
INSERT INTO users (user_id, membership_type, access_token, refresh_token, token_expiry, created_at, updated_at)
VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
);
--

-- name: GetUser :one
SELECT * FROM users WHERE user_id = ?;
--

-- name: UpdateToken :exec
UPDATE users
SET access_token = ?,
    updated_at = ?,
    refresh_token = ?,
    token_expiry = ?
WHERE user_id = ?;