-- +migrate Up
CREATE TABLE users (
    id VARCHAR(100) NOT NULL PRIMARY KEY,
    role VARCHAR(20) NOT NULL
);

-- +migrate Down
DROP TABLE users;