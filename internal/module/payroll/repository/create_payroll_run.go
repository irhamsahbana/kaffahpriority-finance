package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
)

func (r *payrollRepo) CreatePayrollRun(ctx context.Context, run *entity.PayrollRun) error {
	ctx, span := tracing.StartSpan(ctx, "repo.CreatePayrollRun")
	defer span.End()

	query := `
		INSERT INTO payroll_runs (
			id, period_start, period_end, period, timezone, status, created_at
		) VALUES (
			:id, :period_start, :period_end, :period, :timezone, :status, :created_at
		)
	`
	_, err := r.db.NamedExecContext(ctx, query, run)
	return err
}
