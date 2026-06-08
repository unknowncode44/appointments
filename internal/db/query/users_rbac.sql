-- name: ListUsers :many
SELECT id, username, hashed_password, full_name, password_changed_at, created_at, role, tenant_id
FROM users
LIMIT $1 OFFSET $2;

-- name: GetUserByIDFull :one
SELECT id, username, hashed_password, full_name, password_changed_at, created_at, role, tenant_id
FROM users
WHERE id = $1;

-- name: GetUserByUsernameFull :one
SELECT id, username, hashed_password, full_name, password_changed_at, created_at, role, tenant_id
FROM users
WHERE username = $1;

-- name: CreateUserWithRole :one
INSERT INTO users (
  username,
  hashed_password,
  full_name,
  role,
  tenant_id
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING id, username, hashed_password, full_name, password_changed_at, created_at, role, tenant_id;

-- name: UpdateUserRole :one
UPDATE users
SET role = $2, tenant_id = $3
WHERE id = $1
RETURNING id, username, hashed_password, full_name, password_changed_at, created_at, role, tenant_id;

-- name: UpdateUserTenant :one
UPDATE users
SET tenant_id = $2
WHERE id = $1
RETURNING id, username, hashed_password, full_name, password_changed_at, created_at, role, tenant_id;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: AddUserProvider :exec
INSERT INTO user_providers (user_id, provider_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveUserProvider :exec
DELETE FROM user_providers
WHERE user_id = $1 AND provider_id = $2;

-- name: ListUserProviders :many
SELECT provider_id FROM user_providers
WHERE user_id = $1;

