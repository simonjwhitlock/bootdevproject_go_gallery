-- name: GetUserByEmail :one
SELECT id, created_at, updated_at, email
FROM users
WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email)
VALUES ($1, $2, $3, $4)
RETURNING id, created_at, updated_at, email;
