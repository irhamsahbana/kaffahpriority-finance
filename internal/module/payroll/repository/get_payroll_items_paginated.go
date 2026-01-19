package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
	"fmt"
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
		FROM
			payroll_items pi
		WHERE
			pi.payroll_run_id = ?
			AND pi.deleted_at IS NULL
	`
	args = append(args, payrollRunID)

	if req.Q != "" {
		query += ` AND (
			pi.lecturer_name ILIKE '%' || ? || '%' OR
			pi.student_name ILIKE '%' || ? || '%' OR
			pi.program_name ILIKE '%' || ? || '%' OR
			pi.academic_manager_name ILIKE '%' || ? || '%' OR
			pi.marketer_name ILIKE '%' || ? || '%'
		)`
		args = append(args, req.Q, req.Q, req.Q, req.Q, req.Q)
	}

	if req.MarketerID != "" {
		query += ` AND pi.marketer_id = ?`
		args = append(args, req.MarketerID)
	}

	if req.LecturerID != "" {
		query += ` AND pi.lecturer_id = ?`
		args = append(args, req.LecturerID)
	}

	if req.StudentID != "" {
		query += ` AND pi.student_id = ?`
		args = append(args, req.StudentID)
	}

	if req.ProgramID != "" {
		query += ` AND pi.program_id = ?`
		args = append(args, req.ProgramID)
	}

	query += ` ORDER BY prt.created_at ASC LIMIT ? OFFSET ?`
	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	if err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...); err != nil {
		return nil, 0, fmt.Errorf("failed to get payroll items: %w", err)
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
