-- +goose Up
-- +goose StatementBegin
ALTER TABLE program_registration_templates
    ADD COLUMN is_ssp BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN pc_fee DECIMAL(19, 4) NOT NULL DEFAULT 0,
    ADD COLUMN mt_fee DECIMAL(19, 4) NOT NULL DEFAULT 0,
    ADD COLUMN cl_fee DECIMAL(19, 4) NOT NULL DEFAULT 0,
    ADD COLUMN ms_fee DECIMAL(19, 4) NOT NULL DEFAULT 0,
    ADD COLUMN sc_fee DECIMAL(19, 4) NOT NULL DEFAULT 0,
    ADD COLUMN ln_fee DECIMAL(19, 4) NOT NULL DEFAULT 0,
    ADD CONSTRAINT program_registration_templates_no_itp_ssp CHECK (NOT (is_itp AND is_ssp));

ALTER TABLE program_registrations
    ADD COLUMN is_ssp BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN pc_fee DECIMAL(19, 4) NOT NULL DEFAULT 0,
    ADD COLUMN mt_fee DECIMAL(19, 4) NOT NULL DEFAULT 0,
    ADD COLUMN cl_fee DECIMAL(19, 4) NOT NULL DEFAULT 0,
    ADD COLUMN ms_fee DECIMAL(19, 4) NOT NULL DEFAULT 0,
    ADD COLUMN sc_fee DECIMAL(19, 4) NOT NULL DEFAULT 0,
    ADD COLUMN ln_fee DECIMAL(19, 4) NOT NULL DEFAULT 0,
    ADD CONSTRAINT program_registrations_no_itp_ssp CHECK (NOT (is_itp AND is_ssp));

ALTER TABLE payroll_items
    ADD COLUMN is_ssp BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN pc_fee DECIMAL(19, 4) NOT NULL DEFAULT 0,
    ADD COLUMN mt_fee DECIMAL(19, 4) NOT NULL DEFAULT 0,
    ADD COLUMN cl_fee DECIMAL(19, 4) NOT NULL DEFAULT 0,
    ADD COLUMN ms_fee DECIMAL(19, 4) NOT NULL DEFAULT 0,
    ADD COLUMN sc_fee DECIMAL(19, 4) NOT NULL DEFAULT 0,
    ADD COLUMN ln_fee DECIMAL(19, 4) NOT NULL DEFAULT 0,
    ADD CONSTRAINT payroll_items_no_itp_ssp CHECK (NOT (is_itp AND is_ssp));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payroll_items
    DROP CONSTRAINT payroll_items_no_itp_ssp,
    DROP COLUMN ln_fee,
    DROP COLUMN sc_fee,
    DROP COLUMN ms_fee,
    DROP COLUMN cl_fee,
    DROP COLUMN mt_fee,
    DROP COLUMN pc_fee,
    DROP COLUMN is_ssp;

ALTER TABLE program_registrations
    DROP CONSTRAINT program_registrations_no_itp_ssp,
    DROP COLUMN ln_fee,
    DROP COLUMN sc_fee,
    DROP COLUMN ms_fee,
    DROP COLUMN cl_fee,
    DROP COLUMN mt_fee,
    DROP COLUMN pc_fee,
    DROP COLUMN is_ssp;

ALTER TABLE program_registration_templates
    DROP CONSTRAINT program_registration_templates_no_itp_ssp,
    DROP COLUMN ln_fee,
    DROP COLUMN sc_fee,
    DROP COLUMN ms_fee,
    DROP COLUMN cl_fee,
    DROP COLUMN mt_fee,
    DROP COLUMN pc_fee,
    DROP COLUMN is_ssp;
-- +goose StatementEnd
