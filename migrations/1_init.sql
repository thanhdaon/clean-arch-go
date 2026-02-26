-- +migrate Up
CREATE TABLE tasks (
    id VARCHAR(255) NOT NULL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_by VARCHAR(255) NOT NULL,
    assigned_to VARCHAR(255),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT NULL,
    deleted_at DATETIME DEFAULT NULL,
    archived_at DATETIME DEFAULT NULL,
    priority INT DEFAULT 2,
    due_date DATETIME NULL,
    description TEXT NULL
);

CREATE TABLE users (
    id VARCHAR(100) NOT NULL PRIMARY KEY,
    role VARCHAR(20) NOT NULL,
    name VARCHAR(100),
    email VARCHAR(100),
    password_hash VARCHAR(255),
    deleted_at TIMESTAMP NULL
);

CREATE TABLE task_tags (
    id VARCHAR(255) NOT NULL PRIMARY KEY,
    task_id VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- +migrate Down
DROP TABLE task_tags;
DROP TABLE users;
DROP TABLE tasks;
