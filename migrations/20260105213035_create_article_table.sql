-- +goose Up
-- +goose StatementBegin
CREATE TABLE articles (
    uuid UUID PRIMARY KEY,
    external_id VARCHAR(65) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ NOT NULL,
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    content TEXT,
    author VARCHAR(255),
    tags VARCHAR(50)[] DEFAULT '{}',  -- Array of tags
    source SMALLINT NOT NULL,
    
    UNIQUE(external_id, source)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS articles;
-- +goose StatementEnd
