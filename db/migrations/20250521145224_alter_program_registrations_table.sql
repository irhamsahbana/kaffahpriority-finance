-- +goose Up
-- +goose StatementBegin
ALTER TABLE program_registrations
ADD COLUMN allocated_at TIMESTAMP WITH TIME ZONE DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE program_registrations
DROP COLUMN allocated_at;
-- +goose StatementEnd
