-- +goose Up
CREATE TABLE images (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    image_name TEXT NOT NULL,
    image_url TEXT NOT NULL,
    thumbnail_url TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id),
    image_description TEXT
);

-- +goose Down
DROP TABLE images;