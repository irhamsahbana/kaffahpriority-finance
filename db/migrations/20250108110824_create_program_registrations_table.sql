-- Active: 1736401832474@@127.0.0.1@5432@kp_finance_dev
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS program_registrations (
    id CHAR(26) PRIMARY KEY,
    template_id CHAR(26) NOT NULL,
    user_id CHAR(26) NOT NULL,
    program_id CHAR(26) NOT NULL,
    lecturer_id CHAR(26),
    marketer_id CHAR(26) NOT NULL,
    student_id CHAR(26) NOT NULL,
    program_name VARCHAR(255) NOT NULL,
    program_fee DECIMAL(19, 4) NOT NULL,
    program_meetings INT NOT NULL DEFAULT 0,
    program_meetings_completed INT NOT NULL DEFAULT 0,
    administration_fee DECIMAL(19, 4),
    foreign_learning_fee DECIMAL(19, 4),
    night_learning_fee DECIMAL(19, 4),
    marketer_commission_fee DECIMAL(19, 4) NOT NULL DEFAULT 0,
    overpayment_fee DECIMAL(19, 4),
    hr_fee DECIMAL(19, 4) DEFAULT 0 NOT NULL,
    marketer_gifts_fee DECIMAL(19, 4) NOT NULL DEFAULT 0,
    closing_fee_for_office DECIMAL(19, 4),
    closing_fee_for_reward DECIMAL(19, 4),
    days INT[] NOT NULL DEFAULT '{}',
    -- cfo 2 section
    mentor_detail_fee DECIMAL(19, 4),
    hr_detail_fee DECIMAL(19, 4),
    mentor_detail_fee_used DECIMAL(19, 4),
    notes_for_fund_distributions VARCHAR(255),
    used_at TIMESTAMP WITH TIME ZONE,
    -- general section
    notes VARCHAR(255),
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    paid_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,

    FOREIGN KEY (template_id) REFERENCES program_registration_templates (id),
    FOREIGN KEY (program_id) REFERENCES programs (id),
    FOREIGN KEY (lecturer_id) REFERENCES lecturers (id),
    FOREIGN KEY (marketer_id) REFERENCES marketers (id),
    FOREIGN KEY (student_id) REFERENCES students (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS program_registrations;
-- +goose StatementEnd
