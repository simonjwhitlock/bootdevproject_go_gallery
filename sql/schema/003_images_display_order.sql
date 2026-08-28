-- +goose Up
ALTER TABLE images ADD COLUMN display_order INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE images DROP COLUMN display_order;
