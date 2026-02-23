-- +migrate Up
ALTER TABLE tasks ADD COLUMN deleted_at DATETIME DEFAULT NULL;

-- +migrate Down
ALTER TABLE tasks DROP COLUMN deleted_at;
