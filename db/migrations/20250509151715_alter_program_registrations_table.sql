-- +goose Up
-- +goose StatementBegin
ALTER TABLE program_registrations
ADD COLUMN is_paid BOOLEAN DEFAULT FALSE,
ADD COLUMN batch CHAR(26) DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE program_registrations
DROP COLUMN is_paid,
DROP COLUMN batch;
-- +goose StatementEnd
