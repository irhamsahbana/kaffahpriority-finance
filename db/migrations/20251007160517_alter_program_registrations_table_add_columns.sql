-- +goose Up
-- +goose StatementBegin
ALTER TABLE program_registrations
ADD COLUMN parent_id CHAR(26) DEFAULT NULL,
ADD COLUMN category VARCHAR(255) DEFAULT 'general', -- general, additional, shortfall
ADD COLUMN notes_for_category VARCHAR(255) DEFAULT NULL,

ADD CONSTRAINT fk_program_registrations_parent_id
FOREIGN KEY (parent_id)
REFERENCES program_registrations(id)
ON DELETE CASCADE
;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE program_registrations
DROP CONSTRAINT fk_program_registrations_parent_id;

ALTER TABLE program_registrations
DROP COLUMN notes_for_category,
DROP COLUMN category,
DROP COLUMN parent_id;
-- +goose StatementEnd
