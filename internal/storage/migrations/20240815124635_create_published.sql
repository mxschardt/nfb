-- +goose Up
-- +goose StatementBegin
CREATE TABLE posted_articles (
    article_id SERIAL references articles (id),
    channel_id TEXT NOT NULL,
    posted_at TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE articles
    DROP COLUMN posted_at;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS posted_articles;

ALTER TABLE articles
    ADD COLUMN posted_at TIMESTAMP;
-- +goose StatementEnd
