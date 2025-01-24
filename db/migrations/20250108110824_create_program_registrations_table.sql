-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS program_registrations (
    id CHAR(26) PRIMARY KEY,
    user_id CHAR(26) NOT NULL,
    program_id CHAR(26) NOT NULL,
    lecturer_id CHAR(26) NOT NULL,
    marketer_id CHAR(26) NOT NULL,
    student_id CHAR(26) NOT NULL,
    program_name VARCHAR(255) NOT NULL,
    program_fee DECIMAL(19, 4) NOT NULL,
    program_meetings INT NOT NULL,
    program_meetings_completed INT DEFAULT 0 NOT NULL,
    administration_fee DECIMAL(19, 4),
    foreign_learning_fee DECIMAL(19, 4),
    night_learning_fee DECIMAL(19, 4),
    marketer_commission_fee DECIMAL(19, 4) DEFAULT 0 NOT NULL,
    overpayment_fee DECIMAL(19, 4),
    hr_fee DECIMAL(19, 4) DEFAULT 0 NOT NULL,
    marketer_gifts_fee DECIMAL(19, 4) DEFAULT 0 NOT NULL,
    closing_fee_for_office DECIMAL(19, 4),
    closing_fee_for_reward DECIMAL(19, 4),
    days INT[] DEFAULT '{}' NOT NULL,
    -- cfo 2 section
    mentor_detail_fee DECIMAL(19, 4),
    hr_detail_fee DECIMAL(19, 4),
    used_at TIMESTAMP WITH TIME ZONE,
    -- general section
    notes VARCHAR(255),
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    paid_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,

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
