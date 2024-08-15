-- +goose Up
-- +goose StatementBegin
ALTER TABLE articles
    ALTER COLUMN title TYPE TEXT,
    ALTER COLUMN link TYPE TEXT;

ALTER TABLE sources
    ALTER COLUMN name TYPE TEXT,
    ALTER COLUMN feed_url TYPE TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE articles
    ALTER COLUMN title TYPE VARCHAR(255),
    ALTER COLUMN link TYPE VARCHAR(255);

ALTER TABLE sources
    ALTER COLUMN name TYPE VARCHAR(255),
    ALTER COLUMN feed_url TYPE VARCHAR(255);
-- +goose StatementEnd
