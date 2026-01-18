package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
)

func (r *payrollRepo) GetPayrollItem(ctx context.Context, id string) (*entity.PayrollItem, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetPayrollItem")
	defer span.End()

	var item entity.PayrollItem
	query := `SELECT * FROM payroll_items WHERE id = $1 AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
}
