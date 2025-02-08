-- +goose Up
CREATE TABLE IF NOT EXISTS connectors (
    id TEXT PRIMARY KEY,
    workspace_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    default_channel_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS connectors;
