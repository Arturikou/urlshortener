-- migrations/000003_add_column_url_table.up.sql
-- Add column is_deleted to url table

ALTER TABLE url ADD COLUMN is_deleted BOOLEAN DEFAULT FALSE;