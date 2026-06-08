-- users.sql

-- name: CreateUser :one
INSERT INTO users (
  username,
  hashed_password,
  full_name
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: GetUserByUsername :one
SELECT id, username, hashed_password
FROM users
WHERE username = $1;

-- name: UpdateUser :one
UPDATE users
SET
  full_name = $1
WHERE
  username = $2
RETURNING *;

-- name: UpdateUserPassword :one
UPDATE users
SET hashed_password = $1,
    password_changed_at = now()
WHERE username = $2
RETURNING *;
