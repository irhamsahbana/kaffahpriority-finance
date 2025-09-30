package repository

import (
	"codebase-app/internal/module/report/entity"
	"context"

	// "github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

func (r *reportRepo) GetExportedRegistrationsForCFO2Monthly(ctx context.Context, req *entity.GetExportedRegistrationsForCFO2MonthlyReq) (
	*entity.GetExportedRegistrationsForCFO2MonthlyResp, error) {
	var (
		fnName          = "repo::GetExportedRegistrationsForCFO2Monthly"
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
			CASE
				WHEN pr.is_itp THEN pr.program_acquisition_rights * 2
				ELSE pr.program_acquisition_rights
			END AS acquisition_rights,
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
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to fetch data", fnName)
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

	// query, args, err = sqlx.In(query, registrationIds)
	// if err != nil {
	// 	log.Error().Err(err).Any("req", req).Msgf("%s - failed to build query", fnName)
	// 	return nil, err
	// }

	// query = r.db.Rebind(query)
	// err = r.db.SelectContext(ctx, &daosData, query, args...)
	// if err != nil {
	// 	log.Error().Err(err).Any("req", req).Msgf("%s - failed to fetch additional students", fnName)
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

	respUnused.Items = append(respUnused.Items, resp.Items...)
	respUnused.TotalITP += resp.TotalITP

	return respUnused, nil
}
