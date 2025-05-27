package repository

import (
	"codebase-app/internal/module/report/entity"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

func (r *reportRepo) GetRegistrations(ctx context.Context, req *entity.GetRegistrationsReq) (*entity.GetRegistrationsResp, error) {
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
			pr.allocated_at,
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
			am.name AS academic_manager_name,
			sm.name AS student_manager_name
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
		WHERE
			pr.deleted_at IS NULL
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

	if req.LecturerId != "" {
		query += ` AND pr.lecturer_id = ?`
		args = append(args, req.LecturerId)
	}

	if req.MarketerId != "" {
		query += ` AND pr.marketer_id = ?`
		args = append(args, req.MarketerId)
	}

	if req.StudentId != "" {
		query += ` AND pr.student_id = ?`
		args = append(args, req.StudentId)
	}

	if req.ProgramId != "" {
		query += ` AND pr.program_id = ?`
		args = append(args, req.ProgramId)
	}

	if req.IsPaid != "" {
		if req.IsPaid == "true" {
			query += ` AND pr.is_paid = TRUE`
		} else {
			query += ` AND pr.is_paid = FALSE`
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

	query += ` ORDER BY pr.is_paid ASC, ` + sortByMap[req.SortBy] + ` ` + sortTypeMap[req.SortType]
	query += ` LIMIT ? OFFSET ?`
	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetRegistrations - failed to fetch data")
		return nil, err
	}

	for _, item := range data {
		resp.Meta.TotalData = item.TotalData
		registrationIds = append(registrationIds, item.Id)
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
			log.Error().Err(err).Any("req", req).Msg("repo::GetRegistrations - failed to build query")
			return nil, err
		}

		query = r.db.Rebind(query)
		err = r.db.SelectContext(ctx, &daosData, query, args...)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msg("repo::GetRegistrations - failed to fetch additional students")
			return nil, err
		}

		for i, item := range resp.Items {
			for _, data := range daosData {
				if item.Id == data.PrId {
					resp.Items[i].Students = append(resp.Items[i].Students, data.AddStudent)
				}
			}
		}
	}

	return resp, nil
}

func (r *reportRepo) GetExportedRegistrations(ctx context.Context, req *entity.GetExportedRegistrationsReq) (*entity.GetExportedRegistrationsResp, error) {
	var (
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
			s.name AS student_name
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

	if req.LecturerId != "" {
		query += ` AND pr.lecturer_id = ?`
		args = append(args, req.LecturerId)
	}

	if req.MarketerId != "" {
		query += ` AND pr.marketer_id = ?`
		args = append(args, req.MarketerId)
	}

	if req.StudentId != "" {
		query += ` AND pr.student_id = ?`
		args = append(args, req.StudentId)
	}

	if req.ProgramId != "" {
		query += ` AND pr.program_id = ?`
		args = append(args, req.ProgramId)
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

	query += ` ORDER BY ` + sortByMap[req.SortBy] + ` ` + sortTypeMap[req.SortType]

	err := r.db.SelectContext(ctx, &resp.Items, r.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetExportedRegistrations - failed to fetch data")
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

		registrationIds = append(registrationIds, resp.Items[i].Id)
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
	reqSummary.UserId = req.UserId

	respSummary, err := r.GetSummaries(ctx, reqSummary)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetExportedRegistrations - failed to get summaries")
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

	query, args, err = sqlx.In(query, registrationIds)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetExportedRegistrations - failed to build query")
		return nil, err
	}

	query = r.db.Rebind(query)
	err = r.db.SelectContext(ctx, &daosData, query, args...)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetExportedRegistrations - failed to fetch additional students")
		return nil, err
	}

	for i, item := range resp.Items {
		for _, data := range daosData {
			if item.Id == data.PrId {
				resp.Items[i].Students = append(resp.Items[i].Students, data.AddStudent)
				resp.Items[i].StudentName += ", " + *data.AddStudent.Name
			}
		}
	}

	return resp, nil
}

func (r *reportRepo) GetExportedRegistrationsForCFO2Monthly(
	ctx context.Context, req *entity.GetExportedRegistrationsForCFO2MonthlyReq,
) (*entity.GetExportedRegistrationsForCFO2MonthlyResp, error) {
	var (
		resp            = new(entity.GetExportedRegistrationsForCFO2MonthlyResp)
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
			s.name AS student_name
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

	query += ` ORDER BY pr.paid_at ASC`

	err := r.db.SelectContext(ctx, &resp.Items, r.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetExportedRegistrationsForCFO2Monthly - failed to fetch data")
		return nil, err
	}

	respUnused, err := r.GetExportedRegistrationsForCFO2MonthlyUnused(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(resp.Items) == 0 {
		respUnused.Items = append(respUnused.Items, resp.Items...)
		respUnused.TotalITP += resp.TotalITP

		return respUnused, nil
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
			resp.TotalITP++
		}

		registrationIds = append(registrationIds, resp.Items[i].Id)
	}

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

	query, args, err = sqlx.In(query, registrationIds)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetExportedRegistrationsForCFO2Monthly - failed to build query")
		return nil, err
	}

	query = r.db.Rebind(query)
	err = r.db.SelectContext(ctx, &daosData, query, args...)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetExportedRegistrationsForCFO2Monthly - failed to fetch additional students")
		return nil, err
	}

	for i, item := range resp.Items {
		for _, data := range daosData {
			if item.Id == data.PrId {
				resp.Items[i].Students = append(resp.Items[i].Students, data.AddStudent)
				resp.Items[i].StudentName += ", " + *data.AddStudent.Name
			}
		}
	}

	respUnused.Items = append(respUnused.Items, resp.Items...)
	respUnused.TotalITP += resp.TotalITP

	return respUnused, nil
}

func (r *reportRepo) GetExportedRegistrationsForCFO2MonthlyUnused(
	ctx context.Context,
	req *entity.GetExportedRegistrationsForCFO2MonthlyReq) (
	*entity.GetExportedRegistrationsForCFO2MonthlyResp, error) {
	var (
		resp            = new(entity.GetExportedRegistrationsForCFO2MonthlyResp)
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
			s.name AS student_name
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
		WHERE
			pr.deleted_at IS NULL
			AND pr.is_paid = TRUE
			AND pr.mentor_detail_fee_used IS NULL
			AND pr.paid_at AT TIME ZONE ? NOT BETWEEN
				(TO_TIMESTAMP(?, 'YYYY-MM-DD') AT TIME ZONE 'UTC') AND
				(TO_TIMESTAMP(?, 'YYYY-MM-DD') AT TIME ZONE 'UTC' + time '23:59:59.999999')
		ORDER BY pr.paid_at ASC
	`

	err := r.db.SelectContext(ctx, &resp.Items, r.db.Rebind(query), req.Timezone, req.PaidAtFrom, req.PaidAtTo)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetExportedRegistrationsForCFO2MonthlyUnused - failed to fetch data")
		return nil, err
	}

	if len(resp.Items) == 0 {
		return resp, nil
	}

	for i := range resp.Items {
		resp.Items[i].IsUnused = true
		resp.Items[i].Students = make([]entity.AddStudent, 0)

		if resp.Items[i].FLFee != nil {
			resp.Items[i].ProgramName += " + FL"
		}

		if resp.Items[i].NLFee != nil {
			resp.Items[i].ProgramName += " + NL"
		}

		if resp.Items[i].IsITP {
			resp.Items[i].ProgramName += " + ITP"
			resp.TotalITP++
		}

		registrationIds = append(registrationIds, resp.Items[i].Id)
	}

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
		log.Error().Err(err).Any("req", req).Msg("repo::GetExportedRegistrationsForCFO2MonthlyUnused - failed to build query")
		return nil, err
	}

	query = r.db.Rebind(query)
	err = r.db.SelectContext(ctx, &daosData, query, args...)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetExportedRegistrationsForCFO2MonthlyUnused - failed to fetch additional students")
		return nil, err
	}

	for i, item := range resp.Items {
		for _, data := range daosData {
			if item.Id == data.PrId {
				resp.Items[i].Students = append(resp.Items[i].Students, data.AddStudent)
				resp.Items[i].StudentName += ", " + *data.AddStudent.Name
			}
		}
	}

	return resp, err
}

func (r *reportRepo) GetExportedRegistrationsForCFO2Yearly(
	ctx context.Context,
	req *entity.GetExportedRegistrationsForCFO2YearlyReq) (*entity.GetExportedRegistrationsForCFO2YearlyResp, error) {
	type dao struct {
		entity.RegisItem
		PaidAtMonth int     `db:"paid_at_month"`
		Notes       *string `db:"notes"`
	}
	var (
		resp = new(entity.GetExportedRegistrationsForCFO2YearlyResp)
		data = make([]dao, 0)
	)
	resp.Items = make([]entity.RegistrationYearlyRow, 0)

	query := `
		SELECT
			pr.id,
			pr.program_id,
			pr.lecturer_id,
			pr.student_id,
			pr.marketer_id,
			am.id AS academic_manager_id,

			p.name AS program_name,
			l.name AS lecturer_name,
			s.name AS student_name,
			m.name AS marketer_name,
			am.name AS academic_manager_name,

			pr.mentor_detail_fee AS hr_fee_for_mentor,
			TO_CHAR(pr.paid_at AT TIME ZONE ?, 'YYYY-MM-DD HH24:MI:SS') AS paid_at,
			EXTRACT(MONTH FROM pr.paid_at AT TIME ZONE ?) AS paid_at_month,
			notes_for_fund_distributions AS notes
		FROM
			program_registrations pr
		LEFT JOIN
			lecturers l
			ON pr.lecturer_id = l.id
		LEFT JOIN
			academic_managers am
			ON l.academic_manager_id = am.id
		JOIN
			students s
			ON pr.student_id = s.id
		JOIN
			programs p
			ON pr.program_id = p.id
		JOIN
			marketers m
			ON pr.marketer_id = m.id
		WHERE
			pr.deleted_at IS NULL
			AND pr.is_paid = TRUE
			AND pr.lecturer_id IS NOT NULL
			AND pr.paid_at AT TIME ZONE ? BETWEEN
				(TO_TIMESTAMP(?, 'YYYY-MM-DD') AT TIME ZONE 'UTC') AND
				(TO_TIMESTAMP(?, 'YYYY-MM-DD') AT TIME ZONE 'UTC' + time '23:59:59.999999')
		ORDER BY
			am.id ASC,
			l.id ASC,
			pr.id ASC
	`

	paidAtFrom := fmt.Sprintf("%s-01-01", req.PaidAtYear)
	paidAtTo := fmt.Sprintf("%s-12-31", req.PaidAtYear)

	err := r.db.SelectContext(ctx, &data, r.db.Rebind(query),
		req.Timezone,
		req.Timezone,
		req.Timezone,
		paidAtFrom,
		paidAtTo,
	)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetExportedRegistrationsForCFO2Yearly - failed to fetch data")
		return nil, err
	}

	groupExist := make(map[string]struct{})

	ress := make([]entity.RegistrationYearlyRow, 0)

	for _, item := range data {

		key := fmt.Sprintf("%s-%s-%s-%s-%s", *item.AcademicManagerId, *item.LecturerId, item.MarketerId, item.StudentId, item.ProgramId)

		if _, ok := groupExist[key]; !ok {
			groupExist[key] = struct{}{}
			ress = append(ress, entity.RegistrationYearlyRow{
				AcademicManagerId:   *item.AcademicManagerId,
				AcademicManagerName: *item.AcademicManagerName,
				LecturerId:          *item.LecturerId,
				LecturerName:        *item.LecturerName,
				StudentId:           item.StudentId,
				StudentName:         item.StudentName,
				MarketerId:          item.MarketerId,
				MarketerName:        item.MarketerName,
				Months:              []entity.RegistrationYearlyMonth{},
			})

			ress[len(ress)-1].Months = append(ress[len(ress)-1].Months, entity.RegistrationYearlyMonth{
				RegistrationId: item.Id,
				HRFeeForMentor: item.HRFeeForMentor,
				PaidAt:         item.PaidAt,
				PaidAtMonth:    item.PaidAtMonth,
			})

			ress[len(ress)-1].ProgramId = item.ProgramId
			ress[len(ress)-1].ProgramName = item.ProgramName
			ress[len(ress)-1].LecturerId = *item.LecturerId
			ress[len(ress)-1].LecturerName = *item.LecturerName
			ress[len(ress)-1].AcademicManagerId = *item.AcademicManagerId
			ress[len(ress)-1].AcademicManagerName = *item.AcademicManagerName
			ress[len(ress)-1].StudentId = item.StudentId
			ress[len(ress)-1].StudentName = item.StudentName
			ress[len(ress)-1].MarketerId = item.MarketerId
			ress[len(ress)-1].MarketerName = item.MarketerName

		} else {
			for i, res := range ress {
				if res.AcademicManagerId == *item.AcademicManagerId &&
					res.LecturerId == *item.LecturerId &&
					res.MarketerId == item.MarketerId &&
					res.ProgramId == item.ProgramId &&
					res.StudentId == item.StudentId {
					// month in number

					ress[i].Months = append(ress[i].Months, entity.RegistrationYearlyMonth{
						RegistrationId: item.Id,
						HRFeeForMentor: item.HRFeeForMentor,
						PaidAt:         item.PaidAt,
						PaidAtMonth:    item.PaidAtMonth,
						Notes:          item.Notes,
					})
				}
			}
		}
	}

	resp.Items = ress

	return resp, nil
}

func (r *reportRepo) GetExportedRegistrationsForWageRecapMonthly(
	ctx context.Context,
	req *entity.GetExportedRegistrationsForWageRecapMonthlyReq) (*entity.GetExportedRegistrationsForWageRecapMonthlyResp, error) {
	resp := new(entity.GetExportedRegistrationsForWageRecapMonthlyResp)

	queryAcademicManagers := `
		SELECT id, name FROM academic_managers
		WHERE deleted_at IS NULL
	`

	var academicManagers = make([]entity.WageRecapAcademicManager, 0)

	err := r.db.SelectContext(ctx, &academicManagers, r.db.Rebind(queryAcademicManagers))
	if err != nil {
		log.Error().Err(err).Msg("repo::GetExportedRegistrationsForWageRecapMonthly - failed to fetch academic managers")
		return nil, err
	}

	for i := range academicManagers {
		academicManagers[i].Items = make([]entity.WageRecapLecturer, 0)

		query := `
			SELECT
				l.id,
				l.name
			FROM
				lecturers l
			WHERE
				l.deleted_at IS NULL
				AND l.academic_manager_id = ?
		`

		err = r.db.SelectContext(ctx, &academicManagers[i].Items, r.db.Rebind(query), academicManagers[i].Id)
		if err != nil {
			log.Error().Err(err).Msg("repo::GetExportedRegistrationsForWageRecapMonthly - failed to fetch lecturers")
			return nil, err
		}

		for j := range academicManagers[i].Items {
			academicManagers[i].Items[j].Items = make([]entity.WageRecapRegistration, 0)

			query = `
				SELECT
					prt.id,
					prt.student_id,
					prt.program_id,
					prt.lecturer_id,
					s.name AS student_name,
					p.name AS program_name,
					m.name AS marketer_name,

					CASE
						WHEN prt.foreign_learning_fee IS NOT NULL THEN TRUE
						ELSE FALSE
					END AS is_fl,
					CASE
						WHEN prt.night_learning_fee IS NOT NULL THEN TRUE
						ELSE FALSE
					END AS is_nl,
					prt.is_itp
				FROM
					program_registration_templates prt
				JOIN
					students s ON prt.student_id = s.id
				JOIN
					programs p ON prt.program_id = p.id
				JOIN
					marketers m ON prt.marketer_id = m.id
				WHERE
					prt.deleted_at IS NULL
					AND prt.lecturer_id = ?
				ORDER BY
					prt.id ASC
			`

			var registrations = make([]entity.WageRecapRegistration, 0)
			err = r.db.SelectContext(ctx, &registrations, r.db.Rebind(query), academicManagers[i].Items[j].Id)
			if err != nil {
				log.Error().Err(err).Msg("repo::GetExportedRegistrationsForWageRecapMonthly - failed to fetch registrations")
				return nil, err
			}

			academicManagers[i].Items[j].Items = registrations

			for k := range academicManagers[i].Items[j].Items {
				if academicManagers[i].Items[j].Items[k].IsFL {
					academicManagers[i].Items[j].Items[k].ProgramName += " + FL"
				}
				if academicManagers[i].Items[j].Items[k].IsNL {
					academicManagers[i].Items[j].Items[k].ProgramName += " + NL"
				}
				if academicManagers[i].Items[j].Items[k].IsITP {
					academicManagers[i].Items[j].Items[k].ProgramName += " + ITP"
				}

				queryAddStudents := `
					SELECT
						adds.student_id,
						CASE
							WHEN s.id IS NULL THEN adds.name
							ELSE s.name
						END AS name
					FROM
						prt_additional_students adds
					LEFT JOIN
						students s
						ON adds.student_id = s.id
					WHERE
						adds.prt_id = ?
				`

				var addStudents = make([]entity.AddStudent, 0)
				err = r.db.SelectContext(ctx, &addStudents, r.db.Rebind(queryAddStudents), academicManagers[i].Items[j].Items[k].Id)
				if err != nil {
					log.Error().Err(err).Msg("repo::GetExportedRegistrationsForWageRecapMonthly - failed to fetch additional students")
					return nil, err
				}
				for _, addStudent := range addStudents {
					academicManagers[i].Items[j].Items[k].StudentName += ", " + *addStudent.Name
				}

				query := `
					SELECT
						pr.program_meetings, -- jumlah
						pr.program_fee_per_meeting, -- hitungan
						pr.full_fee, -- ujroh full
						pr.is_full_fee,
						COALESCE(pr.initial_fee, (
							CASE
								WHEN pr.initial_fee IS NOT NULL THEN pr.initial_fee
								WHEN pr.is_full_fee = TRUE THEN pr.full_fee
								ELSE pr.program_fee_per_meeting * pr.program_meetings
							END
							)
						) AS initial_fee, -- ujroh awal
						pr.foreign_learning_fee, -- fl
						pr.night_learning_fee, -- nl
						(
							COALESCE(pr.night_learning_fee, 0) +
							COALESCE(pr.foreign_learning_fee, 0) +
							COALESCE(pr.initial_fee,
								CASE
									WHEN pr.is_full_fee THEN pr.full_fee
									ELSE pr.program_fee_per_meeting * pr.program_meetings
								END
							)
						) AS real_fee, -- ujroh real
						CASE
							WHEN pr.is_paid = TRUE THEN pr.mentor_detail_fee_used
							ELSE 0
						END AS mentor_detail_fee_used, -- keep gaji
						CASE
							WHEN pr.is_itp THEN pr.program_acquisition_rights * 2
							ELSE pr.program_acquisition_rights
						END AS acquisition_rights, -- angka
						pr.notes_for_lecturer_wage AS notes
					FROM
						program_registrations pr
					WHERE
						pr.deleted_at IS NULL
						AND pr.is_paid = TRUE
						AND pr.lecturer_id = ?
						AND pr.student_id = ?
						AND pr.program_id = ?
						AND TO_CHAR(pr.allocated_at AT TIME ZONE ?, 'YYYY-MM') = (
							TO_CHAR(
								(((? || '-01')::timestamptz) AT TIME ZONE 'UTC')
								, 'YYYY-MM'
							)
						)
					ORDER BY
						pr.id DESC
					LIMIT 1
				`

				args := make([]any, 0, 5)
				args = append(args,
					academicManagers[i].Items[j].Items[k].LecturerId,
					academicManagers[i].Items[j].Items[k].StudentId,
					academicManagers[i].Items[j].Items[k].ProgramId,
					req.Timezone,
					req.Month,
				)

				var (
					registration entity.WageRecapRegistrationData
				)
				err = r.db.GetContext(ctx, &registration, r.db.Rebind(query), args...)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {

					log.Error().Err(err).Msg("repo::GetExportedRegistrationsForWageRecapMonthly - failed to fetch registration details")
					return nil, err
				}

				if !errors.Is(err, sql.ErrNoRows) {
					academicManagers[i].Items[j].Items[k].Data = &registration
				}
			}
		}

	}

	resp.Items = academicManagers
	return resp, nil
}
