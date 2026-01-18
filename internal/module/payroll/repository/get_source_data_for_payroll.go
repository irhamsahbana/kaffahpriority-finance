package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
)

func (r *payrollRepo) GetSourceDataForPayroll(ctx context.Context, period string, timezone string) ([]entity.PayrollItem, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetSourceDataForPayroll")
	defer span.End()

	var items []entity.PayrollItem

	query := `
		SELECT
			-- generated IDs for PayrollItem will be set in Service
			'' AS id,
			'' AS payroll_run_id,
			COALESCE(am.id, '') AS academic_manager_id,
			COALESCE(am.name, '') AS academic_manager_name,
			l.id AS lecturer_id,
			l.name AS lecturer_name,
			s.id AS student_id,
			s.name AS student_name,
			p.id AS program_id,
			p.name AS program_name,
			m.id AS marketer_id,
			m.name AS marketer_name,
			pr.foreign_learning_fee,
			pr.night_learning_fee,
			pr.is_itp,
			pr.program_meetings,
			pr.is_full_fee AS is_meeting_full,
			
			CASE WHEN pr.is_itp THEN pr.program_fee_per_meeting * 2 ELSE pr.program_fee_per_meeting END AS wage_per_meeting,
			
			CASE WHEN pr.is_itp THEN pr.full_fee * 2 ELSE pr.full_fee END AS full_wage,
			
			(
				CASE
					WHEN pr.program_meetings < 1 THEN 0
					ELSE (
						COALESCE(pr.night_learning_fee, 0) +
						COALESCE(pr.foreign_learning_fee, 0) +
						COALESCE(pr.initial_fee,
							CASE
								WHEN pr.is_full_fee THEN (CASE WHEN pr.is_itp THEN pr.full_fee * 2 ELSE pr.full_fee END)
								ELSE (CASE WHEN pr.is_itp THEN pr.program_fee_per_meeting * 2 ELSE pr.program_fee_per_meeting END) * pr.program_meetings
							END
						)
					)
				END
			) AS wage,
			
			(
				CASE
					WHEN (pr.is_paid = FALSE OR pr.mentor_detail_fee_used IS NULL OR pr.mentor_detail_fee_used = 0) THEN 0
					WHEN COALESCE(pr.hr_detail_fee, 0) <= 0 THEN 0
					ELSE FLOOR(COALESCE(pr.hr_detail_fee, 0) / 40000)
				END
				+
				COALESCE((
					SELECT
						SUM(
							CASE
								WHEN COALESCE(child.hr_detail_fee, 0) <= 0 THEN 0
								ELSE FLOOR(COALESCE(child.hr_detail_fee, 0) / 40000)
							END
						)
					FROM program_registrations child
					WHERE child.parent_id = pr.id
					  AND child.category = 'additional'
					  AND child.deleted_at IS NULL
				), 0)
			)::int AS acquisition_rights,

			NOW() AS created_at,
			NOW() AS updated_at

		FROM
			program_registrations pr
		JOIN
			programs p ON pr.program_id = p.id
		LEFT JOIN
			lecturers l ON pr.lecturer_id = l.id
		LEFT JOIN
			academic_managers am ON l.academic_manager_id = am.id
		JOIN
			students s ON pr.student_id = s.id
		JOIN
			marketers m ON pr.marketer_id = m.id
		WHERE
			pr.deleted_at IS NULL
			AND pr.category = 'general'
			AND TO_CHAR(pr.allocated_at AT TIME ZONE $1, 'YYYY-MM') = $2
	`

	if err := r.db.SelectContext(ctx, &items, query, timezone, period); err != nil {
		return nil, err
	}

	return items, nil
}
