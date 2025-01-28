-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE tasks (
    id VARCHAR(255) NOT NULL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_by VARCHAR(255) NOT NULL,
    assigned_to VARCHAR(255),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT NULL
);


-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE tasks;