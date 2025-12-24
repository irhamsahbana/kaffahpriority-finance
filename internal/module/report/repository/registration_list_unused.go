package repository

import (
	"codebase-app/internal/module/report/entity"
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

func (r *reportRepo) GetUnusedRegistrations(ctx context.Context, req *entity.GetRegistrationsReq) (*entity.GetRegistrationsResp, error) {
	fnName := "repo::GetUnusedRegistrations"
	type dao struct {
		TotalData int `db:"total_data"`
		entity.RegisItem
	}
	var (
		data            = make([]dao, 0, req.Paginate)
		registrationIds = make([]string, 0)
		resp            = new(entity.GetRegistrationsResp)
		args            = make([]any, 0, 3)
	)
	resp.Items = make([]entity.RegisItem, 0)

	query := `
		SELECT
			COUNT(*) OVER() AS total_data,
			pr.batch,
			pr.is_paid,
			pr.category,
			pr.id,
			pr.template_id,
			pr.program_id,
			pr.marketer_id,
			pr.lecturer_id,
			pr.student_id,
			s.identifier AS student_identifier,
			pr.program_name,
			am.id AS academic_manager_id,
			sm.id AS student_manager_id,
			pr.program_fee,
			pr.administration_fee,
			pr.foreign_learning_fee,
			pr.night_learning_fee,
			pr.is_itp,
			pr.marketer_commission_fee,
			pr.overpayment_fee,
			pr.hr_fee,
			pr.mentor_detail_fee AS hr_fee_for_mentor,
			pr.hr_detail_fee AS hr_fee_for_hr,
			pr.mentor_detail_fee - pr.mentor_detail_fee_used AS hr_fee_for_mentor_remaining,
			CASE
				WHEN pr.mentor_detail_fee_used = pr.mentor_detail_fee THEN 'full'
				WHEN pr.hr_fee = 0 THEN 'full'
				WHEN pr.mentor_detail_fee_used IS NULL THEN NULL
				ELSE 'partial'
			END AS hr_fee_for_mentor_status,
			CASE
				WHEN pr.mentor_detail_fee_used IS NOT NULL THEN TRUE
				ELSE FALSE
			END AS is_mentor_detail_fee_used,
			CASE
				WHEN pr.lecturer_id IS NOT NULL THEN TRUE
				ELSE FALSE
			END AS is_mandatory_fields_completed,
			pr.marketer_gifts_fee,
			pr.closing_fee_for_office,
			pr.closing_fee_for_reward,
			pr.paid_at,
			pr.created_at,
			pr.updated_at,
			CASE
				WHEN pr.parent_id IS NOT NULL THEN parent.allocated_at
				ELSE pr.allocated_at
			END AS allocated_at,
			pr.notes,
			pr.notes_for_category,
			pr.program_fee +
			COALESCE(pr.foreign_learning_fee, 0) +
			COALESCE(pr.night_learning_fee, 0) +
			COALESCE(pr.overpayment_fee, 0)
			AS monthly_fee,
			(
				COALESCE(pr.administration_fee, 0)
				+ COALESCE(pr.program_fee, 0)
				+ COALESCE(pr.overpayment_fee, 0)
				+ COALESCE(pr.night_learning_fee, 0)
				+ COALESCE(pr.foreign_learning_fee, 0)
				- COALESCE(pr.marketer_commission_fee, 0)
				- COALESCE(pr.marketer_gifts_fee, 0)
				- COALESCE(pr.hr_fee, 0)
				- COALESCE(pr.overpayment_fee, 0)
				- COALESCE(pr.closing_fee_for_office, 0)
				- COALESCE(pr.closing_fee_for_reward, 0)
			) AS profit,
			l.name AS lecturer_name,
			m.name AS marketer_name,
			s.name AS student_name,
			am.name AS academic_manager_name,
			sm.name AS student_manager_name,
			CASE
				WHEN pr.parent_id IS NOT NULL
				THEN
					CASE
						WHEN parent.program_meetings > 0 THEN TRUE
						ELSE FALSE
					END
				ELSE
					CASE
						WHEN pr.program_meetings > 0 THEN TRUE
						ELSE FALSE
					END
			END AS is_started,
			CASE
				WHEN pr.parent_id IS NOT NULL
				THEN
					CASE
						WHEN COALESCE(pr.hr_detail_fee, 0) <= 0 THEN 0
						ELSE FLOOR(COALESCE(pr.hr_detail_fee, 0) / 40000)
					END
				ELSE
					CASE
						WHEN pr.program_acquisition_rights > 10 THEN 0
						WHEN pr.hr_fee = 0 THEN 0
						WHEN pr.is_itp THEN 2 * pr.program_acquisition_rights
						ELSE pr.program_acquisition_rights
					END
			END AS acquisition_rights
		FROM
			program_registrations pr
		LEFT JOIN
			lecturers l
			ON pr.lecturer_id = l.id
		LEFT JOIN
			academic_managers am
			ON l.academic_manager_id = am.id
		JOIN
			marketers m
			ON pr.marketer_id = m.id
		JOIN
			student_managers sm
			ON m.student_manager_id = sm.id
		JOIN
			students s
			ON pr.student_id = s.id
		JOIN
			programs p
			ON pr.program_id = p.id
		-- join with parent
		LEFT JOIN
			program_registrations parent
			ON pr.parent_id = parent.id
		LEFT JOIN
			students parent_student
			ON parent.student_id = parent_student.id
		WHERE
			pr.deleted_at IS NULL
			AND pr.mentor_detail_fee_used IS NULL
			AND pr.hr_fee > 0
			AND
			parent.deleted_at IS NULL
	`

	if req.PaidAtFrom != "" && req.PaidAtTo != "" {
		query += `
			AND pr.paid_at AT TIME ZONE ? <=
			(TO_TIMESTAMP(?, 'YYYY-MM-DD') AT TIME ZONE 'UTC' - INTERVAL '1 month' + time '23:59:59.999999')
		`
		args = append(args, req.Timezone, req.PaidAtTo)
	}

	if req.IsStarted != "" {
		if req.IsStarted == "true" {
			query += ` AND pr.program_meetings > 0`
		} else {
			query += ` AND pr.program_meetings = 0`
		}
	}

	if req.PaidAtFrom != "" && req.PaidAtTo != "" {
		query += `
			AND CASE
				WHEN pr.parent_id IS NOT NULL
				THEN parent.allocated_at AT TIME ZONE ?
				ELSE pr.allocated_at AT TIME ZONE ?
			END NOT BETWEEN
			(TO_TIMESTAMP(?, 'YYYY-MM-DD') AT TIME ZONE 'UTC') AND
			(TO_TIMESTAMP(?, 'YYYY-MM-DD') AT TIME ZONE 'UTC' - INTERVAL '1 month' + time '23:59:59.999999')`
		args = append(args, req.Timezone, req.Timezone, req.PaidAtFrom, req.PaidAtTo)
	}

	if req.Q != "" {
		query += ` AND (
			s.name ILIKE '%' || ? || '%' OR
			parent_student.name ILIKE '%' || ? || '%'
		)`
		args = append(args, req.Q, req.Q)
	}

	if req.IsMandatoryFieldsCompleted != "" {
		if req.IsMandatoryFieldsCompleted == "true" {
			query += ` AND pr.lecturer_id IS NOT NULL`
		} else {
			query += ` AND pr.lecturer_id IS NULL`
		}
	}

	if req.MentorFeeAllocationStatus != "all" {
		switch req.MentorFeeAllocationStatus {
		case "full":
			query += ` AND (COALESCE(pr.mentor_detail_fee, 0) - COALESCE(pr.mentor_detail_fee_used, 0)) = 0`
		case "partial":
			query += ` AND (COALESCE(pr.mentor_detail_fee, 0) - COALESCE(pr.mentor_detail_fee_used, 0)) > 0`
		}
	}

	if req.LecturerID != "" {
		query += ` AND pr.lecturer_id = ?`
		args = append(args, req.LecturerID)
	}

	if req.MarketerID != "" {
		query += ` AND pr.marketer_id = ?`
		args = append(args, req.MarketerID)
	}

	if req.StudentID != "" {
		query += ` AND pr.student_id = ?`
		args = append(args, req.StudentID)
	}

	if req.ProgramID != "" {
		query += ` AND pr.program_id = ?`
		args = append(args, req.ProgramID)
	}

	if req.IsPaid != "" {
		if req.IsPaid == "true" {
			query += ` AND pr.is_paid = TRUE`
		} else {
			query += ` AND pr.is_paid = FALSE`
		}
	}

	if req.IsITP != "" {
		if req.IsITP == "true" {
			query += ` AND pr.is_itp = TRUE`
		} else {
			query += ` AND pr.is_itp = FALSE`
		}
	}

	sortByMap := map[string]string{
		"created_at":   "pr.created_at",
		"paid_at":      "pr.paid_at",
		"updated_at":   "pr.updated_at",
		"student_name": "s.name",
		"":             "pr.paid_at",
	}

	sortTypeMap := map[string]string{
		"asc":  "ASC",
		"desc": "DESC",
		"":     "DESC",
	}

	if req.SortBy == "paid_at" {
		query += ` ORDER BY pr.is_paid ASC, pr.paid_at ` + sortTypeMap[req.SortType] + `, m.id DESC`
	} else {
		query += ` ORDER BY pr.is_paid ASC, ` + sortByMap[req.SortBy] + ` ` + sortTypeMap[req.SortType]
	}

	query += ` LIMIT ? OFFSET ?`
	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to fetch data", fnName)
		return nil, err
	}

	for _, item := range data {
		resp.Meta.TotalData = item.TotalData
		registrationIds = append(registrationIds, item.ID)
		resp.Items = append(resp.Items, item.RegisItem)
		resp.Items[len(resp.Items)-1].Students = make([]entity.AddStudent, 0)
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	if len(registrationIds) > 0 {
		type daos struct {
			PrId string `db:"pr_id"`
			entity.AddStudent
		}
		var (
			daosData = make([]daos, 0)
		)
		query = `
			SELECT
				prs.pr_id,
				prs.student_id,
				CASE
					WHEN s.id IS NULL THEN prs.name
					ELSE s.name
				END AS name
			FROM
				pr_additional_students prs
			LEFT JOIN
				students s
				ON prs.student_id = s.id
			WHERE prs.pr_id IN (?)
		`

		query, args, err := sqlx.In(query, registrationIds)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msgf("%s - failed to build query", fnName)
			return nil, err
		}

		query = r.db.Rebind(query)
		err = r.db.SelectContext(ctx, &daosData, query, args...)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msgf("%s - failed to fetch additional students", fnName)
			return nil, err
		}

		for i, item := range resp.Items {
			for _, data := range daosData {
				if item.ID == data.PrId {
					resp.Items[i].Students = append(resp.Items[i].Students, data.AddStudent)
				}
			}
		}
	}

	return resp, nil
}
