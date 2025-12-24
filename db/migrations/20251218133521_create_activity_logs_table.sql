-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS activity_logs (
    id CHAR(26) PRIMARY KEY,
    entity_id CHAR(26) NOT NULL,
    entity_name VARCHAR(255),
    logs JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS activity_logs;
-- +goose StatementEnd
