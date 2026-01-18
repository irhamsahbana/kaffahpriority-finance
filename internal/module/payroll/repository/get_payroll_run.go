package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
)

func (r *payrollRepo) GetPayrollRun(ctx context.Context, id string) (*entity.PayrollRun, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetPayrollRun")
	defer span.End()

	var run entity.PayrollRun
	query := `
		SELECT id, period_start, period_end, period, timezone, status, created_at
		FROM payroll_runs
		WHERE id = $1
	`
	if err := r.db.GetContext(ctx, &run, query, id); err != nil {
		return nil, err
	}
	return &run, nil
}
