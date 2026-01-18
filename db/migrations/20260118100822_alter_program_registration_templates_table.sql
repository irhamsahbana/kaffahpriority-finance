-- +goose Up
-- +goose StatementBegin
ALTER TABLE program_registration_templates
ADD COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'active';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE program_registration_templates
DROP COLUMN IF EXISTS status;
-- +goose StatementEnd
