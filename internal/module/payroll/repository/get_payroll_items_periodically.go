package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
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
			COALESCE(SUM(CASE WHEN pi.program_meetings = 0 THEN 0 ELSE (pi.initial_wage + COALESCE(pi.foreign_learning_fee, 0) + COALESCE(pi.night_learning_fee, 0)) END), 0) as wage,
			COALESCE(SUM(CASE WHEN pi.wage > 0 THEN pi.acquisition_rights ELSE 0 END), 0) as total_acquisition_rights
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
			pi.academic_manager_id ASC,
			pi.lecturer_id ASC
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

func (r *payrollRepo) GetPayrollReportsYearly(ctx context.Context, req *entity.GetPayrollReportsYearlyReq) (*entity.PayrollReportsYearlyResp, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetPayrollReportsYearly")
	defer span.End()

	type dao struct {
		TotalData         int    `db:"total_data"`
		AcademicManagerID string `db:"academic_manager_id"`
		entity.PayrollReportsYearlyItem
	}
	type monthData struct {
		Month                  int             `db:"month"`
		Wage                   decimal.Decimal `db:"wage"`
		TotalAcquisitionRights int             `db:"total_acquisition_rights"`
	}

	var (
		resp = new(entity.PayrollReportsYearlyResp)
		data = make([]dao, 0)
		args = make([]any, 0, 3)
	)

	resp.Items = make([]entity.PayrollReportsYearlyItem, 0)

	lecturerQuery := `
		SELECT DISTINCT
			COUNT (*) OVER() AS total_data,
			pi.lecturer_id AS lecturer_id,
			pi.lecturer_name AS lecturer_name,
			pi.academic_manager_id AS academic_manager_id,
			pi.academic_manager_name AS academic_manager_name,
			EXTRACT(YEAR FROM pr.created_at AT TIME ZONE ?) AS year
		FROM
			payroll_items pi
		JOIN
			payroll_runs pr ON pi.payroll_run_id = pr.id
		WHERE
			pi.deleted_at IS NULL
			AND EXTRACT(YEAR FROM pr.created_at AT TIME ZONE ?) = ?
	`
	args = append(args, req.Timezone, req.Timezone, req.Year)

	if req.AcademicManagerID != "" {
		lecturerQuery += ` AND pi.academic_manager_id = ?`
		args = append(args, req.AcademicManagerID)
	}

	if req.LecturerID != "" {
		lecturerQuery += ` AND pi.lecturer_id = ?`
		args = append(args, req.LecturerID)
	}

	lecturerQuery += `
		GROUP BY
			pi.lecturer_id,
			pi.lecturer_name,
			pi.academic_manager_id,
			pi.academic_manager_name,
			year
		ORDER BY
			pi.academic_manager_id ASC,
			pi.lecturer_id ASC
		LIMIT ? OFFSET ?
	`

	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	if err := r.db.SelectContext(ctx, &data, r.db.Rebind(lecturerQuery), args...); err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msg("failed to query payroll reports yearly lecturers")
		return nil, err
	}

	for i, lecturer := range data {
		monthQuery := `
			SELECT
				EXTRACT(MONTH FROM pr.created_at AT TIME ZONE ?) AS month,
				COALESCE(SUM(pi.wage), 0) AS wage,
				COALESCE(SUM(CASE WHEN pi.wage > 0 THEN pi.acquisition_rights ELSE 0 END), 0) AS total_acquisition_rights
			FROM
				payroll_items pi
			JOIN
				payroll_runs pr ON pi.payroll_run_id = pr.id
			WHERE
				pi.deleted_at IS NULL
				AND pi.lecturer_id = ?
				AND EXTRACT(YEAR FROM pr.created_at AT TIME ZONE ?) = ?
		`

		monthArgs := []any{req.Timezone, lecturer.LecturerID, req.Timezone, req.Year}

		if req.AcademicManagerID != "" {
			monthQuery += ` AND pi.academic_manager_id = ?`
			monthArgs = append(monthArgs, req.AcademicManagerID)
		}

		monthQuery += `
			GROUP BY
				month
			ORDER BY
				month ASC
		`

		var monthDataList []monthData
		if err := r.db.SelectContext(ctx, &monthDataList, r.db.Rebind(monthQuery), monthArgs...); err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msg("failed to query payroll reports yearly monthly data")
			return nil, err
		}

		monthMap := make(map[int]monthData)
		for _, month := range monthDataList {
			monthMap[month.Month] = month
		}

		months := make([]entity.PayrollReportsYearlyMonth, 12)
		monthNames := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni",
			"Juli", "Agustus", "September", "Oktober", "November", "Desember"}

		for j := 0; j < 12; j++ {
			monthNum := j + 1
			monthData, exists := monthMap[monthNum]

			months[j] = entity.PayrollReportsYearlyMonth{
				Month:                  monthNames[j],
				Wage:                   decimal.Zero,
				Notes:                  nil,
				TotalAcquisitionRights: 0,
			}

			if exists {
				months[j].Wage = monthData.Wage
				months[j].TotalAcquisitionRights = monthData.TotalAcquisitionRights
			}
		}

		data[i].Months = months
		resp.Meta.TotalData = lecturer.TotalData
	}

	for _, d := range data {
		resp.Items = append(resp.Items, d.PayrollReportsYearlyItem)
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	return resp, nil
}
