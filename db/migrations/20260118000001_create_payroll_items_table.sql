-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS payroll_items (
    id CHAR(26) PRIMARY KEY,
    template_id CHAR(26) NOT NULL,
    payroll_run_id CHAR(26) NOT NULL,
    academic_manager_id CHAR(26) NOT NULL,
    academic_manager_name VARCHAR(255) NOT NULL,
    lecturer_id CHAR(26) NOT NULL,
    lecturer_name VARCHAR(255) NOT NULL,
    student_id CHAR(26) NOT NULL,
    student_name VARCHAR(255) NOT NULL,
    program_id CHAR(26) NOT NULL,
    program_name VARCHAR(255) NOT NULL,
    marketer_id CHAR(26) NOT NULL,
    marketer_name VARCHAR(255) NOT NULL,
    foreign_learning_fee DECIMAL(19,4) DEFAULT 0 NOT NULL,
    night_learning_fee DECIMAL(19,4) DEFAULT 0 NOT NULL,
    is_itp BOOLEAN DEFAULT false NOT NULL,
    program_meetings INTEGER DEFAULT 0 NOT NULL,
    is_meeting_full BOOLEAN DEFAULT false NOT NULL,
    wage_per_meeting DECIMAL(19,4) DEFAULT 0 NOT NULL,
    full_wage DECIMAL(19,4) DEFAULT 0 NOT NULL,
    wage DECIMAL(19,4) DEFAULT 0 NOT NULL,
    acquisition_rights INTEGER DEFAULT 0 NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,

    FOREIGN KEY (payroll_run_id) REFERENCES payroll_runs(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS payroll_item_addtional_students (
    id CHAR(26) PRIMARY KEY,
    payroll_item_id CHAR(26) NOT NULL,
    student_id CHAR(26),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,

    FOREIGN KEY (payroll_item_id) REFERENCES payroll_items (id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES students (id)
);

CREATE INDEX IF NOT EXISTS idx_payroll_items_payroll_run_id ON payroll_items(payroll_run_id);
CREATE INDEX IF NOT EXISTS idx_payroll_item_addtional_students_payroll_item_id ON payroll_item_addtional_students(payroll_item_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS payroll_item_addtional_students;
DROP TABLE IF EXISTS payroll_items;
-- +goose StatementEnd
