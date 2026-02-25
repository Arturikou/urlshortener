-- migrations/000002_create_user_urls_table.up.sql
-- Create user_urls table

CREATE TABLE IF NOT EXISTS user_urls
(
    user_id UUID,
    url_id  INT,
    PRIMARY KEY (user_id, url_id)
);