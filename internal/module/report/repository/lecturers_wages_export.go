package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

func (r *reportRepo) GetExportedLecturersWages(ctx context.Context, req *entity.GetExportedLecturersWagesReq) (*entity.GetExportedLecturersWagesResp, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetExportedLecturersWages")
	defer span.End()

	fnName := "repo::GetExportedLecturersWages"
	type dao struct {
		TotalData int `db:"total_data"`
		entity.LecturersWageItem
	}
	var (
		resp = new(entity.GetExportedLecturersWagesResp)
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
			am.name AS academic_manager_name,
			l.name AS lecturer_name,
			m.name AS marketer_name,
			sm.name AS student_manager_name,
			pr.foreign_learning_fee,
			pr.night_learning_fee,
			pr.is_itp,
			CASE
				WHEN COALESCE(pr.hr_detail_fee, 0) <= 0 THEN 0
				ELSE FLOOR(COALESCE(pr.hr_detail_fee, 0) / 40000)
			END AS acquisition_rights,
			pr.program_meetings,
			CASE WHEN pr.is_itp THEN pr.program_fee_per_meeting * 2 ELSE pr.program_fee_per_meeting END AS program_fee_per_meeting,
			COALESCE(pr.initial_fee, (
				CASE
					WHEN pr.initial_fee IS NOT NULL THEN pr.initial_fee
					WHEN pr.is_full_fee = TRUE THEN (CASE WHEN pr.is_itp THEN pr.full_fee * 2 ELSE pr.full_fee END)
					ELSE (CASE WHEN pr.is_itp THEN pr.program_fee_per_meeting * 2 ELSE pr.program_fee_per_meeting END) * pr.program_meetings
				END
				)
			) AS initial_fee,
			(
				COALESCE(pr.night_learning_fee, 0) +
				COALESCE(pr.foreign_learning_fee, 0) +
				COALESCE(pr.initial_fee,
					CASE
						WHEN pr.is_full_fee THEN (CASE WHEN pr.is_itp THEN pr.full_fee * 2 ELSE pr.full_fee END)
						ELSE (CASE WHEN pr.is_itp THEN pr.program_fee_per_meeting * 2 ELSE pr.program_fee_per_meeting END) * pr.program_meetings
					END
				)
			) AS real_fee,
			pr.is_full_fee,
			CASE WHEN pr.is_itp THEN pr.full_fee * 2 ELSE pr.full_fee END AS full_fee,
			-- pr.mentor_detail_fee_used,
			CASE
				WHEN pr.is_paid = TRUE THEN pr.mentor_detail_fee_used
				ELSE 0
			END AS mentor_detail_fee_used,
			pr.allocated_at,
			pr.notes_for_lecturer_wage AS notes
		FROM
			program_registrations pr
		JOIN
			programs p ON pr.program_id = p.id
		LEFT JOIN
			program_registration_templates prt ON pr.template_id = prt.id
		LEFT JOIN
			lecturers l ON pr.lecturer_id = l.id
		LEFT JOIN
			academic_managers am ON l.academic_manager_id = am.id
		JOIN
			students s ON pr.student_id = s.id
		JOIN
			marketers m ON pr.marketer_id = m.id
		JOIN
			student_managers sm ON m.student_manager_id = sm.id
		WHERE
			pr.deleted_at IS NULL
			AND pr.lecturer_id IS NOT NULL
			AND pr.category = 'general'
			AND TO_CHAR(pr.allocated_at AT TIME ZONE ?, 'YYYY-MM') = ?
	`

	args = append(args, req.Timezone, req.Month)

	if req.AcademicManagerId != "" {
		query += ` AND am.id = ?`
		args = append(args, req.AcademicManagerId)
	}

	if req.LecturerID != "" {
		query += ` AND pr.lecturer_id = ?`
		args = append(args, req.LecturerID)
	}

	if req.IsMandatoryFieldsCompleted != "" {
		switch req.IsMandatoryFieldsCompleted {
		case "true":
			query += ` AND pr.lecturer_id IS NOT NULL`
		case "false":
			query += ` AND pr.lecturer_id IS NULL`
		}
	}

	if req.Q != "" {
		query += ` AND (pr.program_name ILIKE '%' || ? || '%' OR s.name ILIKE '%' || ? || '%' OR l.name ILIKE '%' || ? || '%')`
		args = append(args, req.Q, req.Q, req.Q)
	}

	query += `
		ORDER BY
			l.academic_manager_id ASC,
			l.id ASC,
			COALESCE(prt.created_at, pr.created_at) ASC
	`

	if err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...); err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to query lecturers wages", fnName)
		return nil, err
	}

	registrationIds := make([]string, 0)

	for _, d := range data {
		registrationIds = append(registrationIds, d.RegistrationID)
		if d.FL != nil {
			d.ProgramName = d.ProgramName + " + FL"
		}
		if d.NL != nil {
			d.ProgramName = d.ProgramName + " + NL"
		}
		if d.IsITP {
			d.ProgramName = d.ProgramName + " + ITP"
		}

		resp.Items = append(resp.Items, d.LecturersWageItem)
	}

	if len(registrationIds) == 0 {
		return resp, nil
	}

	// query for additional students
	query = `
		SELECT
			pr.id,
			STRING_AGG(pras.name, ', ' ORDER BY pras.name) AS additional_students
		FROM
			program_registrations pr
		JOIN
			(SELECT DISTINCT pr_id, name FROM pr_additional_students WHERE deleted_at IS NULL) pras ON pr.id = pras.pr_id
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
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to build query for additional students", fnName)
		return nil, err
	}

	query = r.db.Rebind(query)
	err = r.db.SelectContext(ctx, &additionalStudentsData, query, args...)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to fetch additional students", fnName)
		return nil, err
	}

	// Kumpulkan additional students per registration untuk menghindari duplikasi
	additionalStudentsMap := make(map[string]map[string]bool)
	for _, item := range additionalStudentsData {
		if additionalStudentsMap[item.RegistrationId] == nil {
			additionalStudentsMap[item.RegistrationId] = make(map[string]bool)
		}
		// Split string dan tambahkan setiap nama ke map untuk menghilangkan duplikasi
		names := strings.Split(item.AdditionalStudents, ", ")
		for _, name := range names {
			name = strings.TrimSpace(name)
			if name != "" {
				additionalStudentsMap[item.RegistrationId][name] = true
			}
		}
	}

	// Add additional students data to resp.Items field student_name tanpa duplikasi
	for i, d := range resp.Items {
		if nameMap, exists := additionalStudentsMap[d.RegistrationID]; exists && len(nameMap) > 0 {
			names := make([]string, 0, len(nameMap))
			for name := range nameMap {
				names = append(names, name)
			}
			sort.Strings(names)
			resp.Items[i].StudentName += ", " + strings.Join(names, ", ")
		}
	}

	return resp, nil
}
