package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
)

func (r *payrollRepo) GetPayrollRunByPeriod(ctx context.Context, period string) (*entity.PayrollRun, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetPayrollRunByPeriod")
	defer span.End()

	var run entity.PayrollRun
	query := `
		SELECT *
		FROM payroll_runs
		WHERE period = $1 AND deleted_at IS NULL
		LIMIT 1
	`
	if err := r.db.GetContext(ctx, &run, query, period); err != nil {
		return nil, err
	}
	return &run, nil
}
