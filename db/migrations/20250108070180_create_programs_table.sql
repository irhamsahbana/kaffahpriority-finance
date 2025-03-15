-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS programs (
    id CHAR(26) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    detail TEXT,
    price DECIMAL(19, 4) NOT NULL DEFAULT 0,
    price_per_meeting DECIMAL(19, 4) NOT NULL DEFAULT 0,
    full_fee DECIMAL(19, 4) NOT NULL DEFAULT 0, -- ujroh full
    acquisition_rights INT NOT NULL DEFAULT 0,
    days int[] NOT NULL DEFAULT '{}',
    commission_fee DECIMAL(19, 4) NOT NULL DEFAULT 0,
    lecturer_fee DECIMAL(19, 4) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT programs_name_unique UNIQUE (name)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS programs;
-- +goose StatementEnd
