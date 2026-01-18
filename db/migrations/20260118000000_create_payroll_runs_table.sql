-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS payroll_runs (
    id CHAR(26) PRIMARY KEY,
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    period VARCHAR(255) NOT NULL,
    timezone VARCHAR(255) DEFAULT 'Asia/Makassar' NOT NULL,
    status VARCHAR(50) DEFAULT 'draft' NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT payroll_runs_period_unique UNIQUE (period)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS payroll_runs;
-- +goose StatementEnd
