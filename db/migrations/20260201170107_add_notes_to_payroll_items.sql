-- +goose Up
-- +goose StatementBegin
ALTER TABLE payroll_items
ADD COLUMN notes TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payroll_items
DROP COLUMN notes;
-- +goose StatementEnd
