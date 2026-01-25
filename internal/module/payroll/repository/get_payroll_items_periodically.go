package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *payrollRepo) GetPayrollItemsPeriodically(ctx context.Context, req *entity.GetPayrollItemsPeriodicallyReq) (*entity.GetPayrollItemsPeriodicallyResp, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetPayrollItemsPeriodically")
	defer span.End()

	type dao struct {
		TotalData int `db:"total_data"`
		entity.PayrollItemPeriodically
	}

	var (
		data = make([]dao, 0)
		resp = &entity.GetPayrollItemsPeriodicallyResp{
			Items: make([]entity.PayrollItemPeriodically, 0),
		}
		args []interface{}
	)

	query := `
		SELECT
			COUNT(*) OVER() as total_data,
			pi.lecturer_id,
			pi.lecturer_name,
			pi.academic_manager_id,
			pi.academic_manager_name,
			pr.period,
			COALESCE(SUM(pi.wage), 0) as total_real_fee,
			COALESCE(SUM(pi.acquisition_rights), 0) as total_acquisition_rights
		FROM payroll_items pi
		JOIN payroll_runs pr ON pi.payroll_run_id = pr.id
		WHERE pi.deleted_at IS NULL
	`

	if req.Period != "" {
		query += " AND pr.period = ?"
		args = append(args, req.Period)
	}

	if req.Q != "" {
		query += " AND (pi.lecturer_name ILIKE ? OR pi.student_name ILIKE ? OR pi.program_name ILIKE ?)"
		q := "%" + req.Q + "%"
		args = append(args, q, q, q)
	}

	if req.MarketerID != "" {
		query += " AND pi.marketer_id = ?"
		args = append(args, req.MarketerID)
	}

	if req.LecturerID != "" {
		query += " AND pi.lecturer_id = ?"
		args = append(args, req.LecturerID)
	}

	if req.StudentID != "" {
		query += " AND pi.student_id = ?"
		args = append(args, req.StudentID)
	}

	if req.ProgramID != "" {
		query += " AND pi.program_id = ?"
		args = append(args, req.ProgramID)
	}

	if req.AcademicManagerID != "" {
		query += " AND pi.academic_manager_id = ?"
		args = append(args, req.AcademicManagerID)
	}

	query += `
		GROUP BY
			pi.lecturer_id,
			pi.lecturer_name,
			pi.academic_manager_id,
			pi.academic_manager_name,
			pr.period
		ORDER BY
			pi.academic_manager_name ASC,
			pi.lecturer_name ASC
		LIMIT ? OFFSET ?
	`

	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	if err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("repo.GetPayrollItemsPeriodically - SelectContext")
		return nil, err
	}

	for _, d := range data {
		resp.Items = append(resp.Items, d.PayrollItemPeriodically)
		resp.Meta.TotalData = d.TotalData
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	return resp, nil
}
