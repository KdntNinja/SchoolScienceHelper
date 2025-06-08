-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    email TEXT NOT NULL,
    username TEXT,
    password TEXT
);
-- +goose Down
DROP TABLE IF EXISTS users;
