-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS program_registration_templates (
    id CHAR(26) PRIMARY KEY,
    user_id CHAR(26) NOT NULL,
    program_id CHAR(26) NOT NULL,
    lecturer_id CHAR(26) NOT NULL,
    marketer_id CHAR(26) NOT NULL,
    student_id CHAR(26) NOT NULL,
    program_fee DECIMAL(19, 4) NOT NULL,
    administration_fee DECIMAL(19, 4) NOT NULL,
    foreign_lecturer_fee DECIMAL(19, 4),
    night_learning_fee DECIMAL(19, 4),
    marketer_commision_fee DECIMAL(19, 4) DEFAULT 0 NOT NULL,
    overpayment_fee DECIMAL(19, 4),
    hr_fee DECIMAL(19, 4) DEFAULT 0 NOT NULL,
    marketer_gifts_fee DECIMAL(19, 4) DEFAULT 0 NOT NULL,
    closing_fee_for_office DECIMAL(19, 4),
    closing_fee_for_reward DECIMAL(19, 4),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,

    FOREIGN KEY (program_id) REFERENCES programs (id),
    FOREIGN KEY (lecturer_id) REFERENCES lecturers (id),
    FOREIGN KEY (marketer_id) REFERENCES users (id),
    FOREIGN KEY (student_id) REFERENCES students (id),
    FOREIGN KEY (user_id) REFERENCES users (id),

    CONSTRAINT prt_fk_unique UNIQUE (program_id, lecturer_id, marketer_id, student_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS program_registration_templates;
-- +goose StatementEnd
