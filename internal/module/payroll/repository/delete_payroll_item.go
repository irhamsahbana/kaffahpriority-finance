package repository

import (
	"codebase-app/internal/infrastructure/tracing"
	"context"
)

func (r *payrollRepo) DeletePayrollItem(ctx context.Context, id string) error {
	ctx, span := tracing.StartSpan(ctx, "repo.DeletePayrollItem")
	defer span.End()

	query := `UPDATE payroll_items SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
