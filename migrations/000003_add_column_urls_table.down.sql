-- migrations/000003_add_column_url_table.up.sql
-- Rollback column is_deleted

ALTER TABLE url DROP COLUMN is_deleted;