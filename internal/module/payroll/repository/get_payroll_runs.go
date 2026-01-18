package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
)

func (r *payrollRepo) GetPayrollRuns(ctx context.Context, req *entity.GetPayrollRunsReq) ([]entity.PayrollRun, int, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetPayrollRuns")
	defer span.End()

	type dao struct {
		TotalData int `db:"total_data"`
		entity.PayrollRun
	}
	var data []dao
	var args []any

	query := `
		SELECT
			COUNT(*) OVER() as total_data,
			id, period_start, period_end, period, timezone, status, created_at
		FROM payroll_runs
		WHERE 1=1
	`
	if req.Q != "" {
		query += ` AND period ILIKE '%' || ? || '%'`
		args = append(args, req.Q)
	}

	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	if err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...); err != nil {
		return nil, 0, err
	}

	if len(data) == 0 {
		return []entity.PayrollRun{}, 0, nil
	}

	items := make([]entity.PayrollRun, len(data))
	for i, d := range data {
		items[i] = d.PayrollRun
	}

	return items, data[0].TotalData, nil
}
