-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS marketers (
    id CHAR(26) PRIMARY KEY,
    student_manager_id CHAR(26) NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(255),
    registered_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,

    FOREIGN KEY (student_manager_id) REFERENCES student_managers (id),
    CONSTRAINT marketers_email_unique UNIQUE (email),
    CONSTRAINT marketers_phone_unique UNIQUE (phone)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS marketers;
-- +goose StatementEnd
