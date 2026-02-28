-- +migrate Up
CREATE TABLE comments (
    id VARCHAR(255) NOT NULL PRIMARY KEY,
    task_id VARCHAR(255) NOT NULL,
    author_id VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT NULL,
    deleted_at DATETIME DEFAULT NULL,
    INDEX idx_comments_task_id (task_id),
    INDEX idx_comments_author_id (author_id),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE TABLE comment_references (
    id VARCHAR(255) NOT NULL PRIMARY KEY,
    comment_id VARCHAR(255) NOT NULL,
    reference_type VARCHAR(20) NOT NULL,
    reference_id VARCHAR(255) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_refs_comment_id (comment_id),
    INDEX idx_refs_type_id (reference_type, reference_id),
    FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE
);

CREATE TABLE task_activities (
    id VARCHAR(255) NOT NULL PRIMARY KEY,
    task_id VARCHAR(255) NOT NULL,
    actor_id VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    payload JSON,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_activities_task_id_created (task_id, created_at DESC)
);

-- +migrate Down
DROP TABLE task_activities;
DROP TABLE comment_references;
DROP TABLE comments;
