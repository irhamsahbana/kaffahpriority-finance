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

func (r *reportRepo) GetExportedRegistrationsForCFO2Monthly(ctx context.Context, req *entity.GetExportedRegistrationsForCFO2MonthlyReq) (
	*entity.GetExportedRegistrationsForCFO2MonthlyResp, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetExportedRegistrationsForCFO2Monthly")
	defer span.End()

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
				WHEN pr.mentor_detail_fee_used = pr.mentor_detail_fee THEN 'full'
				WHEN pr.mentor_detail_fee_used IS NULL OR pr.mentor_detail_fee_used = 0 THEN NULL
				ELSE 'partial'
			END AS hr_fee_for_mentor_status,
			CASE
				WHEN COALESCE(pr.mentor_detail_fee_used, 0) > 0 THEN TRUE
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
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to fetch data", fnName)
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

		// Kumpulkan additional students per registration untuk menghindari duplikasi
		additionalStudentsMap := make(map[string]map[string]bool)
		additionalStudentsList := make(map[string][]entity.AddStudent)
		for _, item := range additionalStudentsData {
			if additionalStudentsMap[item.RegistrationId] == nil {
				additionalStudentsMap[item.RegistrationId] = make(map[string]bool)
				additionalStudentsList[item.RegistrationId] = make([]entity.AddStudent, 0)
			}
			// Split string dan tambahkan setiap nama ke map untuk menghilangkan duplikasi
			names := strings.Split(item.AdditionalStudents, ", ")
			for _, name := range names {
				name = strings.TrimSpace(name)
				if name != "" && !additionalStudentsMap[item.RegistrationId][name] {
					additionalStudentsMap[item.RegistrationId][name] = true
					// Create AddStudent entity for Students array
					namePtr := name
					additionalStudentsList[item.RegistrationId] = append(additionalStudentsList[item.RegistrationId], entity.AddStudent{
						Name: &namePtr,
					})
				}
			}
		}

		// Add additional students to resp.Items
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
	}

	respUnused.Items = append(respUnused.Items, resp.Items...)
	respUnused.TotalITP += resp.TotalITP

	return respUnused, nil
}
