-- +migrate Up
ALTER TABLE users ADD COLUMN name VARCHAR(100);
ALTER TABLE users ADD COLUMN email VARCHAR(100);
ALTER TABLE users ADD COLUMN password_hash VARCHAR(255);
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP NULL;

-- +migrate Down
ALTER TABLE users DROP COLUMN deleted_at;
ALTER TABLE users DROP COLUMN password_hash;
ALTER TABLE users DROP COLUMN email;
ALTER TABLE users DROP COLUMN name;
