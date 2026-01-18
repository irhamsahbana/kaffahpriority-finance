package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
)

func (r *payrollRepo) GetPayrollItemsWithPagination(ctx context.Context, payrollRunID string, req *entity.GetPayrollItemsReq) ([]entity.PayrollItem, int, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetPayrollItemsWithPagination")
	defer span.End()

	type dao struct {
		TotalData int `db:"total_data"`
		entity.PayrollItem
	}
	var data []dao
	var args []any

	query := `
		SELECT
			COUNT(*) OVER() as total_data,
			pi.*
		FROM payroll_items pi
		WHERE pi.payroll_run_id = ? AND pi.deleted_at IS NULL
	`
	args = append(args, payrollRunID)

	query += ` ORDER BY pi.created_at ASC LIMIT ? OFFSET ?`
	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	if err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...); err != nil {
		return nil, 0, err
	}

	if len(data) == 0 {
		return []entity.PayrollItem{}, 0, nil
	}

	items := make([]entity.PayrollItem, len(data))
	for i, d := range data {
		items[i] = d.PayrollItem
	}

	return items, data[0].TotalData, nil
}
