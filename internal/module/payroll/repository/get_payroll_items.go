package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
)

func (r *payrollRepo) GetPayrollItems(ctx context.Context, payrollRunID string) ([]entity.PayrollItem, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetPayrollItems")
	defer span.End()

	var items []entity.PayrollItem
	query := `
		SELECT *
		FROM payroll_items
		WHERE payroll_run_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
	`
	if err := r.db.SelectContext(ctx, &items, query, payrollRunID); err != nil {
		return nil, err
	}
	return items, nil
}
