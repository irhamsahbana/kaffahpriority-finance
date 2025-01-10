-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS pr_additional_students (
    id CHAR(26) PRIMARY KEY,
    pr_id CHAR(26) NOT NULL,
    student_id CHAR(26),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,

    FOREIGN KEY (pr_id) REFERENCES program_registrations (id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES students (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pr_additional_students;
-- +goose StatementEnd
