-- name: ListImages :many
SELECT id, created_at, updated_at, image_name, image_url, thumbnail_url, user_id, image_description, display_order
FROM images
ORDER BY display_order ASC, created_at DESC;

-- name: GetImage :one
SELECT id, created_at, updated_at, image_name, image_url, thumbnail_url, user_id, image_description, display_order
FROM images
WHERE id = $1;

-- name: CreateImage :one
INSERT INTO images (id, created_at, updated_at, image_name, image_url, thumbnail_url, user_id, image_description, display_order)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, created_at, updated_at, image_name, image_url, thumbnail_url, user_id, image_description, display_order;

-- name: UpdateImage :one
UPDATE images
SET image_name = $2, image_url = $3, thumbnail_url = $4, image_description = $5, updated_at = $6, display_order = $7
WHERE id = $1
RETURNING id, created_at, updated_at, image_name, image_url, thumbnail_url, user_id, image_description, display_order;

-- name: DeleteImage :exec
DELETE FROM images WHERE id = $1;

-- name: ReorderImages :exec
UPDATE images
SET display_order = $2, updated_at = $3
WHERE id = $1;
