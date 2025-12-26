package repository

import (
	"codebase-app/internal/entity"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

func (r *reportRepo) GetRegistrations(ctx context.Context, req *entity.GetRegistrationsReq) (*entity.GetRegistrationsResp, error) {
	fnName := "repo::GetRegistrations"
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
				WHEN pr.mentor_detail_fee_used >= pr.mentor_detail_fee THEN 'full'
				WHEN pr.hr_fee = 0 THEN 'full'
				WHEN pr.mentor_detail_fee_used IS NULL OR pr.mentor_detail_fee_used = 0 THEN NULL
				ELSE 'partial'
			END AS hr_fee_for_mentor_status,
			CASE
				WHEN pr.mentor_detail_fee_used IS NOT NULL OR pr.mentor_detail_fee_used > 0 THEN TRUE
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
				WHEN
					pr.allocated_at IS NULL AND 
					pr.parent_id IS NOT NULL 
					THEN parent.allocated_at
				ELSE pr.allocated_at
			END AS allocated_at,
			pr.notes,
			pr.notes_for_category,
			pr.program_fee +
			COALESCE(pr.administration_fee, 0) +
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
				WHEN COALESCE(pr.hr_detail_fee, 0) <= 0 THEN 0
				ELSE FLOOR(COALESCE(pr.hr_detail_fee, 0) / 40000)
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
			AND
			parent.deleted_at IS NULL
	`

	if req.PaidAtFrom != "" && req.PaidAtTo != "" {
		query += `
			AND pr.paid_at AT TIME ZONE ? >=
			(TO_TIMESTAMP(?, 'YYYY-MM-DD') AT TIME ZONE 'UTC')
			AND pr.paid_at AT TIME ZONE ? <
			(TO_TIMESTAMP(?, 'YYYY-MM-DD') AT TIME ZONE 'UTC' + INTERVAL '1 day')`
		args = append(args, req.Timezone, req.PaidAtFrom, req.Timezone, req.PaidAtTo)
	}

	// Filter khusus untuk CFO2: exclude registrasi dengan allocated_at di luar range dan mentor_detail_fee_used masih NULL
	// if req.IsCFO2 == "true" && req.PaidAtFrom != "" && req.PaidAtTo != "" {
	// 	query += `
	// 		AND NOT (
	// 			CASE
	// 				WHEN pr.parent_id IS NOT NULL
	// 				THEN parent.allocated_at AT TIME ZONE ?
	// 				ELSE pr.allocated_at AT TIME ZONE ?
	// 			END NOT BETWEEN
	// 			(TO_TIMESTAMP(?, 'YYYY-MM-DD') AT TIME ZONE 'UTC') AND
	// 			(TO_TIMESTAMP(?, 'YYYY-MM-DD') AT TIME ZONE 'UTC' + time '23:59:59.999999')
	// 			AND pr.mentor_detail_fee_used IS NULL
	// 		)`
	// 	args = append(args, req.Timezone, req.Timezone, req.PaidAtFrom, req.PaidAtTo)
	// }

	if req.IsStarted != "" {
		if req.IsStarted == "true" {
			query += ` AND pr.program_meetings > 0`
		} else {
			query += ` AND pr.program_meetings = 0`
		}
	}

	if req.Q != "" {
		query += ` AND (
			s.name ILIKE '%' || ? || '%' OR
			parent_student.name ILIKE '%' || ? || '%'
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
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to fetch data", fnName)
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
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to build query", fnName)
			return nil, err
		}

		query = r.db.Rebind(query)
		err = r.db.SelectContext(ctx, &daosData, query, args...)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to fetch additional students", fnName)
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

func (r *reportRepo) GetExportedRegistrationsForCFO2MonthlyUnused(
	ctx context.Context,
	req *entity.GetExportedRegistrationsForCFO2MonthlyReq) (
	*entity.GetExportedRegistrationsForCFO2MonthlyResp, error) {
	var (
		fnName          = "repo::GetExportedRegistrationsForCFO2MonthlyUnused"
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
			CASE
				WHEN COALESCE(pr.hr_detail_fee, 0) <= 0 THEN 0
				ELSE FLOOR(COALESCE(pr.hr_detail_fee, 0) / 40000)
			END AS acquisition_rights,
			pr.marketer_commission_fee,
			pr.overpayment_fee,
			pr.hr_fee,
			pr.mentor_detail_fee AS hr_fee_for_mentor,
			pr.hr_detail_fee AS hr_fee_for_hr,
			pr.mentor_detail_fee - pr.mentor_detail_fee_used AS hr_fee_for_mentor_remaining,
			CASE
				WHEN pr.mentor_detail_fee_used >= pr.mentor_detail_fee THEN 'full'
				WHEN pr.hr_fee = 0 THEN 'full'
				WHEN pr.mentor_detail_fee_used IS NULL OR pr.mentor_detail_fee_used = 0 THEN NULL
				ELSE 'partial'
			END AS hr_fee_for_mentor_status,
			CASE
				WHEN pr.mentor_detail_fee_used IS NOT NULL OR pr.mentor_detail_fee_used > 0 THEN TRUE
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
			COALESCE(pr.administration_fee, 0) +
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
			AND pr.hr_fee > 0
			AND pr.is_paid = TRUE
			AND (pr.mentor_detail_fee_used IS NULL OR pr.mentor_detail_fee_used = 0)
			AND pr.paid_at AT TIME ZONE ? NOT BETWEEN
				(TO_TIMESTAMP(?, 'YYYY-MM-DD') AT TIME ZONE 'UTC') AND
				(TO_TIMESTAMP(?, 'YYYY-MM-DD') AT TIME ZONE 'UTC' + time '23:59:59.999999')
		ORDER BY pr.paid_at ASC
	`

	err := r.db.SelectContext(ctx, &resp.Items, r.db.Rebind(query), req.Timezone, req.PaidAtFrom, req.PaidAtTo)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to fetch data", fnName)
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

		registrationIds = append(registrationIds, resp.Items[i].ID)
	}

	// type daos struct {
	// 	PrId string `db:"pr_id"`
	// 	entity.AddStudent
	// }

	// var (
	// 	daosData = make([]daos, 0)
	// )

	// query = `
	// 		SELECT
	// 			prs.pr_id,
	// 			prs.student_id,
	// 			CASE
	// 				WHEN s.id IS NULL THEN prs.name
	// 				ELSE s.name
	// 			END AS name
	// 		FROM
	// 			pr_additional_students prs
	// 		LEFT JOIN
	// 			students s
	// 			ON prs.student_id = s.id
	// 		WHERE prs.pr_id IN (?)
	// 	`

	// query, args, err := sqlx.In(query, registrationIds)
	// if err != nil {
	// 	log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to build query", fnName)
	// 	return nil, err
	// }

	// query = r.db.Rebind(query)
	// err = r.db.SelectContext(ctx, &daosData, query, args...)
	// if err != nil {
	// 	log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to fetch additional students", fnName)
	// 	return nil, err
	// }

	// for i, item := range resp.Items {
	// 	for _, data := range daosData {
	// 		if item.ID == data.PrId {
	// 			resp.Items[i].Students = append(resp.Items[i].Students, data.AddStudent)
	// 			resp.Items[i].StudentName += ", " + *data.AddStudent.Name
	// 		}
	// 	}
	// }

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
		fnName = "repo::GetExportedRegistrationsForCFO2Yearly"
		resp   = new(entity.GetExportedRegistrationsForCFO2YearlyResp)
		data   = make([]dao, 0)
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
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to fetch data", fnName)
		return nil, err
	}

	groupExist := make(map[string]struct{})

	ress := make([]entity.RegistrationYearlyRow, 0)

	for _, item := range data {

		key := fmt.Sprintf("%s-%s-%s-%s-%s", *item.AcademicManagerID, *item.LecturerID, item.MarketerID, item.StudentID, item.ProgramID)

		if _, ok := groupExist[key]; !ok {
			groupExist[key] = struct{}{}
			ress = append(ress, entity.RegistrationYearlyRow{
				AcademicManagerID:   *item.AcademicManagerID,
				AcademicManagerName: *item.AcademicManagerName,
				LecturerID:          *item.LecturerID,
				LecturerName:        *item.LecturerName,
				StudentID:           item.StudentID,
				StudentName:         item.StudentName,
				MarketerID:          item.MarketerID,
				MarketerName:        item.MarketerName,
				Months:              []entity.RegistrationYearlyMonth{},
			})

			ress[len(ress)-1].Months = append(ress[len(ress)-1].Months, entity.RegistrationYearlyMonth{
				RegistrationID: item.ID,
				HRFeeForMentor: item.HRFeeForMentor,
				PaidAt:         item.PaidAt,
				PaidAtMonth:    item.PaidAtMonth,
			})

			ress[len(ress)-1].ProgramID = item.ProgramID
			ress[len(ress)-1].ProgramName = item.ProgramName
			ress[len(ress)-1].LecturerID = *item.LecturerID
			ress[len(ress)-1].LecturerName = *item.LecturerName
			ress[len(ress)-1].AcademicManagerID = *item.AcademicManagerID
			ress[len(ress)-1].AcademicManagerName = *item.AcademicManagerName
			ress[len(ress)-1].StudentID = item.StudentID
			ress[len(ress)-1].StudentName = item.StudentName
			ress[len(ress)-1].MarketerID = item.MarketerID
			ress[len(ress)-1].MarketerName = item.MarketerName

		} else {
			for i, res := range ress {
				if res.AcademicManagerID == *item.AcademicManagerID &&
					res.LecturerID == *item.LecturerID &&
					res.MarketerID == item.MarketerID &&
					res.ProgramID == item.ProgramID &&
					res.StudentID == item.StudentID {
					// month in number

					ress[i].Months = append(ress[i].Months, entity.RegistrationYearlyMonth{
						RegistrationID: item.ID,
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

	// Collect all registration IDs for additional students query
	registrationIds := make([]string, 0)
	for _, item := range data {
		registrationIds = append(registrationIds, item.ID)
	}

	// Query additional students if there are registrations
	if len(registrationIds) > 0 {
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

		var additionalStudentsData = make([]additionalStudents, 0)

		query, args, err := sqlx.In(query, registrationIds)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - error preparing query for additional students", fnName)
			return nil, err
		}

		err = r.db.SelectContext(ctx, &additionalStudentsData, r.db.Rebind(query), args...)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to fetch additional students", fnName)
			return nil, err
		}

		// Create map of registration ID to additional students (deduplicated)
		additionalStudentsMap := make(map[string]map[string]bool)
		for _, item := range additionalStudentsData {
			if additionalStudentsMap[item.RegistrationId] == nil {
				additionalStudentsMap[item.RegistrationId] = make(map[string]bool)
			}
			// Split string and add each name to map to remove duplicates
			names := strings.Split(item.AdditionalStudents, ", ")
			for _, name := range names {
				name = strings.TrimSpace(name)
				if name != "" {
					additionalStudentsMap[item.RegistrationId][name] = true
				}
			}
		}

		// Add additional students to StudentName for each row
		for i := range resp.Items {
			// Collect all unique additional students for all registrations in this row
			allAdditionalStudents := make(map[string]bool)
			for _, month := range resp.Items[i].Months {
				if nameMap, exists := additionalStudentsMap[month.RegistrationID]; exists {
					for name := range nameMap {
						allAdditionalStudents[name] = true
					}
				}
			}

			// Add to StudentName without duplicates
			if len(allAdditionalStudents) > 0 {
				names := make([]string, 0, len(allAdditionalStudents))
				for name := range allAdditionalStudents {
					names = append(names, name)
				}
				sort.Strings(names)
				resp.Items[i].StudentName += ", " + strings.Join(names, ", ")
			}
		}
	}

	return resp, nil
}

func (r *reportRepo) GetExportedRegistrationsForWageRecapMonthly(
	ctx context.Context,
	req *entity.GetExportedRegistrationsForWageRecapMonthlyReq) (*entity.GetExportedRegistrationsForWageRecapMonthlyResp, error) {
	var (
		fnName = "repo::GetExportedRegistrationsForWageRecapMonthly"
		resp   = new(entity.GetExportedRegistrationsForWageRecapMonthlyResp)
	)

	queryAcademicManagers := `
		SELECT DISTINCT
			am.id, am.name
		FROM
			program_registrations pr
		LEFT JOIN
			lecturers l ON pr.lecturer_id = l.id
		LEFT JOIN
			academic_managers am ON l.academic_manager_id = am.id
		WHERE
			pr.deleted_at IS NULL
			AND pr.category = 'general'
			AND pr.lecturer_id IS NOT NULL
			AND TO_CHAR(pr.allocated_at AT TIME ZONE ?, 'YYYY-MM') = ?
		ORDER BY
			am.id ASC
	`

	var academicManagers = make([]entity.WageRecapAcademicManager, 0)

	err := r.db.SelectContext(
		ctx, &academicManagers, r.db.Rebind(queryAcademicManagers),
		req.Timezone, req.Month,
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("%s - failed to fetch academic managers", fnName)
		return nil, err
	}

	for i := range academicManagers { // looping for each academic manager
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

		err = r.db.SelectContext(ctx, &academicManagers[i].Items, r.db.Rebind(query), academicManagers[i].ID)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msgf("%s - failed to fetch lecturers", fnName)
			return nil, err
		}

		for j := range academicManagers[i].Items { // looping for each lecturer
			academicManagers[i].Items[j].Items = make([]entity.WageRecapRegistration, 0)

			query = `
				SELECT
					pr.id,
					pr.student_id,
					pr.program_id,
					pr.lecturer_id,
					s.name AS student_name,
					p.name AS program_name,
					m.name AS marketer_name,

					CASE
						WHEN pr.foreign_learning_fee IS NOT NULL THEN TRUE
						ELSE FALSE
					END AS is_fl,
					CASE
						WHEN pr.night_learning_fee IS NOT NULL THEN TRUE
						ELSE FALSE
					END AS is_nl,
					pr.is_itp
				FROM
					program_registrations pr
				LEFT JOIN
					program_registration_templates prt ON pr.template_id = prt.id
				JOIN
					students s ON pr.student_id = s.id
				JOIN
					programs p ON pr.program_id = p.id
				JOIN
					marketers m ON pr.marketer_id = m.id
				WHERE
					pr.lecturer_id = ?
					AND pr.deleted_at IS NULL
					AND pr.category = 'general'
					AND TO_CHAR(pr.allocated_at AT TIME ZONE ?, 'YYYY-MM') = ?
				ORDER BY
					COALESCE(prt.created_at, pr.created_at) ASC
			`

			var registrations = make([]entity.WageRecapRegistration, 0)
			err = r.db.SelectContext(ctx, &registrations, r.db.Rebind(query),
				academicManagers[i].Items[j].ID,
				req.Timezone,
				req.Month,
			)
			if err != nil {
				log.Ctx(ctx).Error().Err(err).Msgf("%s - failed to fetch registrations", fnName)
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
						pras.student_id,
						CASE
							WHEN s.id IS NULL THEN pras.name
							ELSE s.name
						END AS name
					FROM
						(SELECT DISTINCT pr_id, student_id, name FROM pr_additional_students WHERE deleted_at IS NULL) pras
					LEFT JOIN
						students s
						ON pras.student_id = s.id
					WHERE
						pras.pr_id = ?
					ORDER BY
						CASE
							WHEN s.id IS NULL THEN pras.name
							ELSE s.name
						END
				`

				var addStudents = make([]entity.AddStudent, 0)
				err = r.db.SelectContext(ctx, &addStudents, r.db.Rebind(queryAddStudents), academicManagers[i].Items[j].Items[k].ID)
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Msgf("%s - failed to fetch additional students", fnName)
					return nil, err
				}

				// Kumpulkan nama untuk menghindari duplikasi
				uniqueNames := make(map[string]bool)
				names := make([]string, 0)
				for _, addStudent := range addStudents {
					name := *addStudent.Name
					if !uniqueNames[name] {
						uniqueNames[name] = true
						names = append(names, name)
					}
				}

				// Sort dan gabungkan
				sort.Strings(names)
				if len(names) > 0 {
					academicManagers[i].Items[j].Items[k].StudentName += ", " + strings.Join(names, ", ")
				}

				query := `
					SELECT
						pr.program_meetings, -- jumlah
						CASE WHEN pr.is_itp THEN pr.program_fee_per_meeting * 2 ELSE pr.program_fee_per_meeting END AS program_fee_per_meeting, -- hitungan
						CASE WHEN pr.is_itp THEN pr.full_fee * 2 ELSE pr.full_fee END AS full_fee, -- ujroh full
						pr.is_full_fee,
						COALESCE(pr.initial_fee, (
							CASE
								WHEN pr.initial_fee IS NOT NULL THEN pr.initial_fee
								WHEN pr.is_full_fee = TRUE THEN (CASE WHEN pr.is_itp THEN pr.full_fee * 2 ELSE pr.full_fee END)
								ELSE (CASE WHEN pr.is_itp THEN pr.program_fee_per_meeting * 2 ELSE pr.program_fee_per_meeting END) * pr.program_meetings
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
									WHEN pr.is_full_fee THEN (CASE WHEN pr.is_itp THEN pr.full_fee * 2 ELSE pr.full_fee END)
									ELSE (CASE WHEN pr.is_itp THEN pr.program_fee_per_meeting * 2 ELSE pr.program_fee_per_meeting END) * pr.program_meetings
								END
							)
						) AS real_fee, -- ujroh real
						CASE
							WHEN pr.program_meetings < 1 THEN 0
							WHEN pr.is_paid = TRUE THEN pr.mentor_detail_fee_used
							ELSE 0
						END AS mentor_detail_fee_used, -- keep gaji
						CASE
							WHEN (pr.is_paid = FALSE OR pr.mentor_detail_fee_used IS NULL OR pr.mentor_detail_fee_used = 0) THEN 0
							WHEN COALESCE(pr.hr_detail_fee, 0) <= 0 THEN 0
							ELSE FLOOR(COALESCE(pr.hr_detail_fee, 0) / 40000)
						END AS acquisition_rights, -- angka
						pr.notes_for_lecturer_wage AS notes
					FROM
						program_registrations pr
					WHERE
						pr.deleted_at IS NULL
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
					academicManagers[i].Items[j].Items[k].LecturerID,
					academicManagers[i].Items[j].Items[k].StudentID,
					academicManagers[i].Items[j].Items[k].ProgramID,
					req.Timezone,
					req.Month,
				)

				var (
					registration entity.WageRecapRegistrationData
				)
				err = r.db.GetContext(ctx, &registration, r.db.Rebind(query), args...)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {

					log.Ctx(ctx).Error().Err(err).Msgf("%s - failed to fetch registration details", fnName)
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
