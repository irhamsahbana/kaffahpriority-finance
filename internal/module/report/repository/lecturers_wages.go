package repository

import (
	"codebase-app/internal/module/report/entity"
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

func (r *reportRepo) GetLecturersWages(ctx context.Context, req *entity.GetLecturersWagesReq) (*entity.GetLecturersWagesResp, error) {
	type dao struct {
		TotalData int `db:"total_data"`
		entity.LecturersWageItem
	}
	var (
		resp = new(entity.GetLecturersWagesResp)
		data = make([]dao, 0)
		args = make([]any, 0, 3)
	)

	resp.Items = make([]entity.LecturersWageItem, 0)

	query := `
		SELECT
			COUNT (*) OVER() AS total_data,
			pr.id AS registration_id,
			pr.program_name,
			s.name AS student_name,
			l.name AS lecturer_name,
			m.name AS marketer_name,
			pr.foreign_learning_fee,
			pr.night_learning_fee,
			pr.is_itp,
			CASE
				WHEN pr.is_itp THEN pr.program_acquisition_rights * 2
				ELSE pr.program_acquisition_rights
			END AS acquisition_rights,
			pr.program_meetings,
			pr.program_fee_per_meeting,
			(
				CASE
					WHEN pr.is_full_fee = TRUE THEN pr.full_fee
					ELSE pr.program_fee_per_meeting * pr.program_meetings
				END
			) AS initial_fee,
			(
				(
				CASE
					WHEN pr.is_full_fee = TRUE THEN pr.full_fee
					ELSE pr.program_fee_per_meeting * pr.program_meetings
				END
				) + COALESCE(pr.night_learning_fee, 0) + COALESCE(pr.foreign_learning_fee, 0)
			) AS real_fee,
			pr.is_full_fee,
			pr.full_fee,
			pr.mentor_detail_fee_used,
			pr.notes_for_lecturer_wage AS notes
		FROM
			program_registrations pr
		JOIN
			programs p ON pr.program_id = p.id
		JOIN
			lecturers l ON pr.lecturer_id = l.id
		JOIN
			academic_managers am ON l.academic_manager_id = am.id
		JOIN
			students s ON pr.student_id = s.id
		JOIN
			marketers m ON pr.marketer_id = m.id
		WHERE
			pr.deleted_at IS NULL
			AND TO_CHAR(pr.paid_at AT TIME ZONE ?, 'YYYY-MM') = ?
	`

	args = append(args, req.Timezone, req.Month)

	if req.AcademicManagerId != "" {
		query += ` AND am.id = ?`
		args = append(args, req.AcademicManagerId)
	}

	if req.LecturerId != "" {
		query += ` AND pr.lecturer_id = ?`
		args = append(args, req.LecturerId)
	}

	query += `
		ORDER BY
			pr.lecturer_id ASC,
			pr.student_id ASC,
			pr.program_name ASC
		LIMIT ? OFFSET ?
	`
	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	if err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetLecturersWages - failed to query lecturers wages")
		return nil, err
	}

	registrationIds := make([]string, 0)

	for _, d := range data {
		registrationIds = append(registrationIds, d.RegistrationId)
		resp.Meta.TotalData = d.TotalData
		resp.Items = append(resp.Items, d.LecturersWageItem)
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	if len(registrationIds) == 0 {
		return resp, nil
	}

	// query for additional students
	query = `
		SELECT
			pr.id,
			STRING_AGG(pras.name, ', ') AS additional_students
		FROM
			program_registrations pr
		JOIN
			pr_additional_students pras ON pr.id = pras.pr_id
		WHERE
			pr.id IN (?)
		GROUP BY
			pr.id
	`

	type additionalStudents struct {
		RegistrationId     string `db:"id"`
		AdditionalStudents string `db:"additional_students"`
	}

	var additionalStudentsData []additionalStudents

	query, args, err := sqlx.In(query, registrationIds)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetLecturersWages - failed to build query for additional students")
		return nil, err
	}

	query = r.db.Rebind(query)
	err = r.db.SelectContext(ctx, &additionalStudentsData, query, args...)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetLecturersWages - failed to fetch additional students")
		return nil, err
	}

	// Add additional students data to resp.Items field student_name
	for _, item := range additionalStudentsData {
		for i, d := range resp.Items {
			if d.RegistrationId == item.RegistrationId {
				resp.Items[i].StudentName += ", " + item.AdditionalStudents

				break
			}
		}
	}

	return resp, nil
}
