package repository

import (
	"codebase-app/internal/module/report/entity"
	"context"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

func (r *reportRepo) GetExportedRegistrations(ctx context.Context, req *entity.GetExportedRegistrationsReq) (*entity.GetExportedRegistrationsResp, error) {
	var (
		fnName          = "repo::GetExportedRegistrations"
		resp            = new(entity.GetExportedRegistrationsResp)
		args            = make([]any, 0, 3)
		registrationIds = make([]string, 0)
	)

	resp.Items = make([]entity.RegisItem, 0)

	query := `
		SELECT
			pr.batch,
			pr.is_paid,
			pr.id,
			pr.template_id,
			pr.program_id,
			pr.marketer_id,
			pr.lecturer_id,
			pr.student_id,
			s.identifier AS student_identifier,
			pr.program_name,
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
			pr.allocated_at,
			pr.created_at,
			pr.updated_at,
			pr.notes,
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
			CASE
				WHEN pr.program_meetings > 0 THEN TRUE
				ELSE FALSE
			END AS is_started
		FROM
			program_registrations pr
		LEFT JOIN
			lecturers l
			ON pr.lecturer_id = l.id
		JOIN
			marketers m
			ON pr.marketer_id = m.id
		JOIN
			students s
			ON pr.student_id = s.id
		JOIN
			programs p
			ON pr.program_id = p.id
		WHERE
			pr.deleted_at IS NULL
			AND pr.is_paid = TRUE
	`

	if req.PaidAtFrom != "" && req.PaidAtTo != "" {
		query += `
				AND pr.paid_at AT TIME ZONE ? BETWEEN
				(TO_TIMESTAMP(?, 'YYYY-MM-DD') AT TIME ZONE 'UTC') AND
				(TO_TIMESTAMP(?, 'YYYY-MM-DD') AT TIME ZONE 'UTC' + time '23:59:59.999999')
		`
		args = append(args, req.Timezone, req.PaidAtFrom, req.PaidAtTo)
	}

	if req.Q != "" {
		query += ` AND (
			pr.program_name ILIKE '%' || ? || '%' OR
			s.name ILIKE '%' || ? || '%'
		)`
		args = append(args, req.Q, req.Q)
	}

	if req.IsLecturerFeeUsed != "" {
		if req.IsLecturerFeeUsed == "true" {
			query += ` AND pr.mentor_detail_fee_used IS NOT NULL`
		} else {
			query += ` AND pr.mentor_detail_fee_used IS NULL`
		}
	}

	if req.IsMandatoryFieldsCompleted != "" {
		if req.IsMandatoryFieldsCompleted == "true" {
			query += ` AND pr.lecturer_id IS NOT NULL`
		} else {
			query += ` AND pr.lecturer_id IS NULL`
		}
	}

	if req.MentorFeeAllocationStatus != "all" {
		if req.MentorFeeAllocationStatus == "full" {
			query += ` AND (COALESCE(pr.mentor_detail_fee, 0) - COALESCE(pr.mentor_detail_fee_used, 0)) = 0`
		} else if req.MentorFeeAllocationStatus == "partial" {
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
		"":     "ASC",
	}

	query += ` ORDER BY ` + sortByMap[req.SortBy] + ` ` + sortTypeMap[req.SortType]

	err := r.db.SelectContext(ctx, &resp.Items, r.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to fetch data", fnName)
		return nil, err
	}

	for i := range resp.Items {
		resp.Items[i].Students = make([]entity.AddStudent, 0)

		if resp.Items[i].FLFee != nil {
			resp.Items[i].ProgramName += " + FL"
		}

		if resp.Items[i].NLFee != nil {
			resp.Items[i].ProgramName += " + NL"
		}

		if resp.Items[i].IsITP {
			resp.Items[i].ProgramName += " + ITP"
		}

		registrationIds = append(registrationIds, resp.Items[i].ID)
	}

	type daos struct {
		PrId string `db:"pr_id"`
		entity.AddStudent
	}

	// get summaries
	reqSummary := &entity.GetSummariesReq{}
	reqSummary.PaidAtFrom = req.PaidAtFrom
	reqSummary.PaidAtTo = req.PaidAtTo
	reqSummary.Timezone = req.Timezone
	reqSummary.UserID = req.UserID

	respSummary, err := r.GetSummaries(ctx, reqSummary)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to get summaries", fnName)
		return nil, err
	}

	resp.Summary = respSummary

	// if no data found then return
	if len(resp.Items) == 0 {
		return resp, nil
	}

	var (
		daosData = make([]daos, 0)
	)
	query = `
			SELECT DISTINCT ON (prs.pr_id, 
				CASE
					WHEN s.id IS NULL THEN prs.name
					ELSE s.name
				END)
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
				AND prs.deleted_at IS NULL
			ORDER BY prs.pr_id, 
				CASE
					WHEN s.id IS NULL THEN prs.name
					ELSE s.name
				END
		`

	query, args, err = sqlx.In(query, registrationIds)
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

	// Kumpulkan additional students per registration untuk menghindari duplikasi
	additionalStudentsMap := make(map[string]map[string]bool)
	additionalStudentsList := make(map[string][]entity.AddStudent)
	for _, data := range daosData {
		if additionalStudentsMap[data.PrId] == nil {
			additionalStudentsMap[data.PrId] = make(map[string]bool)
			additionalStudentsList[data.PrId] = make([]entity.AddStudent, 0)
		}
		name := *data.AddStudent.Name
		if !additionalStudentsMap[data.PrId][name] {
			additionalStudentsMap[data.PrId][name] = true
			additionalStudentsList[data.PrId] = append(additionalStudentsList[data.PrId], data.AddStudent)
		}
	}

	for i, item := range resp.Items {
		if nameMap, exists := additionalStudentsMap[item.ID]; exists && len(nameMap) > 0 {
			// Tambahkan ke Students array
			resp.Items[i].Students = append(resp.Items[i].Students, additionalStudentsList[item.ID]...)

			// Tambahkan ke StudentName tanpa duplikasi
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
