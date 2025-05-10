-- +goose Up
-- +goose StatementBegin
ALTER TABLE program_registration_templates
ALTER COLUMN marketer_id DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE program_registration_templates
ALTER COLUMN marketer_id SET NOT NULL;
-- +goose StatementEnd
