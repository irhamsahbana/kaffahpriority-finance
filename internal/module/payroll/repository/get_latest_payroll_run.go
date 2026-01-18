package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
)

func (r *payrollRepo) GetLatestPayrollRun(ctx context.Context) (*entity.PayrollRun, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetLatestPayrollRun")
	defer span.End()

	var run entity.PayrollRun
	query := `
		SELECT *
		FROM payroll_runs
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`
	if err := r.db.GetContext(ctx, &run, query); err != nil {
		return nil, err
	}
	return &run, nil
}
