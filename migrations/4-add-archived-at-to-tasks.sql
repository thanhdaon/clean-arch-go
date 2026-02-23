-- +migrate Up
ALTER TABLE tasks ADD COLUMN archived_at DATETIME DEFAULT NULL;

-- +migrate Down
ALTER TABLE tasks DROP COLUMN archived_at;
