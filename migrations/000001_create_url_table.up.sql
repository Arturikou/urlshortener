-- migrations/000001_create_url_table.up.sql
-- Create url table

CREATE TABLE IF NOT EXISTS url
(
    id    SERIAL PRIMARY KEY,
    url   TEXT NOT NULL UNIQUE,
    alias VARCHAR(255) NOT NULL UNIQUE
);
CREATE INDEX IF NOT EXISTS idx_url_alias ON url(alias);