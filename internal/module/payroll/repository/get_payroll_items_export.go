package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
)

func (r *payrollRepo) GetPayrollItemsForExport(ctx context.Context, payrollRunID string) ([]entity.PayrollItem, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetPayrollItemsForExport")
	defer span.End()

	var items []entity.PayrollItem
	query := `
		SELECT
			p.*
		FROM payroll_items p
		JOIN
			program_registration_templates t
			ON p.template_id = t.id
		WHERE p.payroll_run_id = $1 AND p.deleted_at IS NULL
		ORDER BY p.academic_manager_id ASC, p.lecturer_id ASC, t.created_at ASC
	`
	if err := r.db.SelectContext(ctx, &items, query, payrollRunID); err != nil {
		return nil, err
	}
	return items, nil
}
