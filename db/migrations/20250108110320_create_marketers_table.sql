-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS marketers (
    id CHAR(26) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(255),
    registered_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT marketers_email_unique UNIQUE (email),
    CONSTRAINT marketers_phone_unique UNIQUE (phone)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS marketers;
-- +goose StatementEnd
