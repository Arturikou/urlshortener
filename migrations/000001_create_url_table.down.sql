-- migrations/000001_create_url_table.down.sql
-- Rollback url database

DROP INDEX IF EXISTS idx_url_alias;
DROP TABLE IF EXISTS url CASCADE;