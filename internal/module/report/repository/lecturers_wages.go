package repository

import (
	"codebase-app/internal/module/report/entity"
	"codebase-app/pkg/errmsg"
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

func (r *reportRepo) GetLecturersWages(ctx context.Context, req *entity.GetLecturersWagesReq) (*entity.GetLecturersWagesResp, error) {
	fnName := "repo::GetLecturersWages"
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
			pr.template_id AS template_id,
			pr.program_name,
			s.name AS student_name,
			am.name AS academic_manager_name,
			l.name AS lecturer_name,
			m.name AS marketer_name,
			sm.name AS student_manager_name,
			pr.foreign_learning_fee,
			pr.night_learning_fee,
			pr.is_itp,
			pr.is_paid AS has_payment,
			(
				CASE
					WHEN pr.program_acquisition_rights > 10 THEN 0
					WHEN pr.hr_fee = 0 THEN 0
					WHEN (pr.is_paid = FALSE OR pr.mentor_detail_fee_used IS NULL OR pr.mentor_detail_fee_used = 0) THEN 0
					WHEN pr.is_itp THEN pr.program_acquisition_rights * 2
					ELSE pr.program_acquisition_rights
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
			) AS real_fee,
			pr.is_full_fee,
			CASE WHEN pr.is_itp THEN pr.full_fee * 2 ELSE pr.full_fee END AS full_fee,
			CASE
				WHEN pr.program_meetings < 1 THEN 0
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
			AND pr.category = 'general'
			AND TO_CHAR(pr.allocated_at AT TIME ZONE ?, 'YYYY-MM') = ?
	`

	args = append(args, req.Timezone, req.Month)

	if req.AcademicManagerID != "" {
		query += ` AND am.id = ?`
		args = append(args, req.AcademicManagerID)
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
		LIMIT ? OFFSET ?
	`
	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	if err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...); err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to query lecturers wages", fnName)
		return nil, err
	}

	registrationIds := make([]string, 0)

	for _, d := range data {
		registrationIds = append(registrationIds, d.RegistrationID)
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
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to build query for additional students", fnName)
		return nil, err
	}

	query = r.db.Rebind(query)
	err = r.db.SelectContext(ctx, &additionalStudentsData, query, args...)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to fetch additional students", fnName)
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

func (r *reportRepo) GetLecturersWagesAggregate(ctx context.Context, req *entity.GetLecturersWagesAggregateReq) (*entity.LecturersWageAggregateResp, error) {
	fnName := "repo::GetLecturersWagesAggregate"
	type dao struct {
		TotalData int `db:"total_data"`
		entity.LecturersWageAggregateItem
	}
	var (
		resp = new(entity.LecturersWageAggregateResp)
		data = make([]dao, 0)
		args = make([]any, 0, 3)
	)

	resp.Items = make([]entity.LecturersWageAggregateItem, 0)

	query := `
		SELECT
			COUNT (*) OVER() AS total_data,
			l.id AS lecturer_id,
			l.name AS lecturer_name,
			am.id AS academic_manager_id,
			am.name AS academic_manager_name,
			TO_CHAR(pr.allocated_at AT TIME ZONE ?, 'YYYY-MM') AS month,
			COALESCE(
				SUM(
					(
						COALESCE(pr.night_learning_fee, 0) +
						COALESCE(pr.foreign_learning_fee, 0) +
						COALESCE(pr.initial_fee,
							CASE
								WHEN pr.is_full_fee THEN (CASE WHEN pr.is_itp THEN pr.full_fee * 2 ELSE pr.full_fee END)
								ELSE (CASE WHEN pr.is_itp THEN pr.program_fee_per_meeting * 2 ELSE pr.program_fee_per_meeting END) * pr.program_meetings
							END
						)
					)
				),
				0
			) AS total_real_fee,
			COALESCE(
				SUM(
					CASE
						WHEN pr.program_acquisition_rights > 10 THEN 0
						WHEN (pr.is_paid = FALSE OR pr.mentor_detail_fee_used IS NULL OR pr.mentor_detail_fee_used = 0) THEN 0
						WHEN pr.is_itp THEN pr.program_acquisition_rights * 2
						ELSE pr.program_acquisition_rights
					END
				),
				0
			) AS total_acquisition_rights
		FROM
			program_registrations pr
		JOIN
			lecturers l ON pr.lecturer_id = l.id
		JOIN
			academic_managers am ON l.academic_manager_id = am.id
		WHERE
			pr.deleted_at IS NULL
	`
	args = append(args, req.Timezone)

	if req.Month != "" {
		query += ` AND TO_CHAR(pr.allocated_at AT TIME ZONE ?, 'YYYY-MM') = ?`
		args = append(args, req.Timezone, req.Month)
	}

	if req.AcademicManagerID != "" {
		query += ` AND am.id = ?`
		args = append(args, req.AcademicManagerID)
	}

	if req.LecturerID != "" {
		query += ` AND pr.lecturer_id = ?`
		args = append(args, req.LecturerID)
	}

	query += `
		GROUP BY
			l.id,
			l.name,
			am.id,
			am.name,
			month
		ORDER BY
			l.id ASC,
			month ASC
		LIMIT ? OFFSET ?
	`

	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	if err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...); err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to query lecturers wages", fnName)
		return nil, err
	}

	for _, d := range data {
		resp.Items = append(resp.Items, d.LecturersWageAggregateItem)
		resp.Meta.TotalData = d.TotalData
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	return resp, nil
}

func (r *reportRepo) GetLecturersWagesAggregateYearly(ctx context.Context, req *entity.GetLecturersWagesAggregateYearlyReq) (*entity.LecturersWageAggregateYearlyResp, error) {
	fnName := "repo::GetLecturersWagesAggregateYearly"
	type dao struct {
		TotalData         int    `db:"total_data"`
		AcademicManagerID string `db:"academic_manager_id"`
		entity.LecturersWageAggregateYearlyItem
	}
	type monthData struct {
		Month                  int             `db:"month"`
		UsedAmount             decimal.Decimal `db:"used_amount"`
		TotalAcquisitionRights int             `db:"total_acquisition_rights"`
	}
	var (
		resp = new(entity.LecturersWageAggregateYearlyResp)
		data = make([]dao, 0)
		args = make([]any, 0, 3)
	)

	resp.Items = make([]entity.LecturersWageAggregateYearlyItem, 0)

	// First, get all lecturers for the year
	lecturerQuery := `
		SELECT DISTINCT
			COUNT (*) OVER() AS total_data,
			l.id AS lecturer_id,
			l.name AS lecturer_name,
			am.id AS academic_manager_id,
			am.name AS academic_manager_name,
			EXTRACT(YEAR FROM pr.allocated_at AT TIME ZONE ?) AS year
		FROM
			program_registrations pr
		JOIN
			lecturers l ON pr.lecturer_id = l.id
		JOIN
			academic_managers am ON l.academic_manager_id = am.id
		WHERE
			pr.deleted_at IS NULL
			AND EXTRACT(YEAR FROM pr.allocated_at AT TIME ZONE ?) = ?
	`
	args = append(args, req.Timezone, req.Timezone, req.Year)

	if req.AcademicManagerID != "" {
		lecturerQuery += ` AND am.id = ?`
		args = append(args, req.AcademicManagerID)
	}

	if req.LecturerID != "" {
		lecturerQuery += ` AND pr.lecturer_id = ?`
		args = append(args, req.LecturerID)
	}

	lecturerQuery += `
		GROUP BY
			l.id,
			l.name,
			am.id,
			am.name,
			year
		ORDER BY
			am.id ASC,
			l.id ASC
		LIMIT ? OFFSET ?
	`

	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	if err := r.db.SelectContext(ctx, &data, r.db.Rebind(lecturerQuery), args...); err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to query lecturers", fnName)
		return nil, err
	}

	// For each lecturer, get monthly data
	for i, lecturer := range data {
		monthQuery := `
			SELECT
				EXTRACT(MONTH FROM pr.allocated_at AT TIME ZONE ?) AS month,
				COALESCE(
					SUM(
						(
							COALESCE(pr.night_learning_fee, 0) +
							COALESCE(pr.foreign_learning_fee, 0) +
							COALESCE(pr.initial_fee,
								CASE
									WHEN pr.is_full_fee THEN pr.full_fee
									ELSE pr.program_fee_per_meeting * pr.program_meetings
								END
							)
						)
					),
					0
				) AS used_amount,
				COALESCE(
					SUM(
						CASE
							WHEN pr.program_acquisition_rights > 10 THEN 0
							WHEN (pr.is_paid = FALSE OR pr.mentor_detail_fee_used IS NULL OR pr.mentor_detail_fee_used = 0) THEN 0
							WHEN pr.is_itp THEN pr.program_acquisition_rights * 2
							ELSE pr.program_acquisition_rights
						END
					),
					0
				) AS total_acquisition_rights
			FROM
				program_registrations pr
			WHERE
				pr.deleted_at IS NULL
				AND pr.lecturer_id = ?
				AND EXTRACT(YEAR FROM pr.allocated_at AT TIME ZONE ?) = ?
			GROUP BY
				month
			ORDER BY
				month ASC
		`

		var monthDataList []monthData
		if err := r.db.SelectContext(ctx, &monthDataList, r.db.Rebind(monthQuery), req.Timezone, lecturer.LecturerID, req.Timezone, req.Year); err != nil {
			log.Error().Err(err).Any("req", req).Msgf("%s - failed to query monthly data for lecturer %s", fnName, lecturer.LecturerID)
			return nil, err
		}

		// Create month map for easy lookup
		monthMap := make(map[int]monthData)
		for _, month := range monthDataList {
			monthMap[month.Month] = month
		}

		// Create months array with all 12 months
		months := make([]entity.LecturersWageMonthItem, 12)
		monthNames := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni",
			"Juli", "Agustus", "September", "Oktober", "November", "Desember"}

		for j := 0; j < 12; j++ {
			monthNum := j + 1
			monthData, exists := monthMap[monthNum]

			months[j] = entity.LecturersWageMonthItem{
				Month:                  monthNames[j],
				UsedAmount:             decimal.Zero,
				Notes:                  nil,
				TotalAcquisitionRights: 0,
			}

			if exists {
				months[j].UsedAmount = monthData.UsedAmount
				months[j].TotalAcquisitionRights = monthData.TotalAcquisitionRights
			}
		}

		// Update the lecturer data with months
		data[i].Months = months
		resp.Meta.TotalData = lecturer.TotalData
	}

	// Convert to response format
	for _, d := range data {
		resp.Items = append(resp.Items, d.LecturersWageAggregateYearlyItem)
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	return resp, nil
}

func (r *reportRepo) UpdateLecturersWage(ctx context.Context, req *entity.UpdateLecturersWageReq) error {
	fnName := "repo::UpdateLecturersWage"
	queryParts := []string{}
	args := []any{}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to begin transaction", fnName)
		return err
	}
	defer tx.Rollback()

	if req.ProgramMeetings.Present && req.ProgramMeetings.Valid {
		queryParts = append(queryParts, "program_meetings = ?")
		args = append(args, req.ProgramMeetings.Val)
	}

	if req.InitialFee.Present {
		if req.InitialFee.Valid {
			queryParts = append(queryParts, "initial_fee = ?")
			args = append(args, req.InitialFee.Val)
		} else {
			queryParts = append(queryParts, "initial_fee = NULL")
		}
	}

	if req.FL.Present {
		if req.FL.Valid {
			queryParts = append(queryParts, "foreign_learning_fee = ?")
			args = append(args, req.FL.Val)
		} else {
			queryParts = append(queryParts, "foreign_learning_fee = NULL")
		}
	}

	if req.NL.Present {
		if req.NL.Valid {
			queryParts = append(queryParts, "night_learning_fee = ?")
			args = append(args, req.NL.Val)
		} else {
			queryParts = append(queryParts, "night_learning_fee = NULL")
		}
	}

	if req.IsFullFee.Present && req.IsFullFee.Valid {
		queryParts = append(queryParts, "is_full_fee = ?")
		args = append(args, req.IsFullFee.Val)
	}

	if req.Notes.Present {
		if req.Notes.Valid {
			queryParts = append(queryParts, "notes_for_lecturer_wage = ?")
			args = append(args, req.Notes.Val)
		} else {
			queryParts = append(queryParts, "notes_for_lecturer_wage = NULL")
		}
	}

	// Jika tidak ada field yang berubah, langsung return
	if len(queryParts) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
			UPDATE program_registrations
			SET %s
			WHERE id = ?
		`,
		strings.Join(queryParts, ", "),
	)
	args = append(args, req.RegistrationID)

	_, err = r.db.ExecContext(ctx, r.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to update lecturers wages", fnName)
		return err
	}

	// jika ada perubahan, update juga used_amount menggunakan ujroh real
	// 1. cari kalkulasi real fee
	realFee, err := r.GetRealFee(ctx, tx, req.RegistrationID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warn().Err(err).Any("req", req).Msgf("%s - real fee not found", fnName)
			return errmsg.NewCustomErrors(404).SetMessage("Laporan tidak ditemukan")
		}
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to get real fee", fnName)
		return err
	}

	// 2. update used_amount menggunakan real fee
	err = r.UpdateUsedAmountWithRealFee(ctx, tx, req.RegistrationID, realFee)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to update used amount with real fee", fnName)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to commit transaction", fnName)
		return err
	}

	return nil
}

func (r *reportRepo) BulkUpdateLecturersWage(ctx context.Context, reqs *entity.BulkUpdateLecturersWageReq) error {
	fnName := "repo::BulkUpdateLecturersWage"

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Any("req", reqs).Msgf("%s - failed to begin transaction", fnName)
		return err
	}
	defer tx.Rollback()

	// 1. update lecturers wages
	for _, req := range reqs.Data {
		queryParts := []string{}
		args := []any{}

		if req.InitialFee.Present {
			if req.InitialFee.Valid {
				queryParts = append(queryParts, "initial_fee = ?")
				args = append(args, req.InitialFee.Val)
			} else {
				queryParts = append(queryParts, "initial_fee = NULL")
			}
		}

		if req.IsFullFee.Present && req.IsFullFee.Valid {
			queryParts = append(queryParts, "is_full_fee = ?")
			args = append(args, req.IsFullFee.Val)
		}

		if len(queryParts) == 0 {
			continue
		}

		query := fmt.Sprintf(`
			UPDATE program_registrations
			SET %s
			WHERE id = ?
		`,
			strings.Join(queryParts, ", "),
		)
		args = append(args, req.RegistrationID)

		_, err = r.db.ExecContext(ctx, r.db.Rebind(query), args...)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msgf("%s - failed to update lecturers wages", fnName)
			return err
		}

		// 2. update used_amount menggunakan ujroh real
		realFee, err := r.GetRealFee(ctx, tx, req.RegistrationID)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Warn().Err(err).Any("req", req).Msgf("%s - real fee not found", fnName)
				return errmsg.NewCustomErrors(404).SetMessage("Laporan tidak ditemukan")
			}
			log.Error().Err(err).Any("req", req).Msgf("%s - failed to get real fee", fnName)
			return err
		}

		// 3. update used_amount menggunakan real fee
		err = r.UpdateUsedAmountWithRealFee(ctx, tx, req.RegistrationID, realFee)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msgf("%s - failed to update used amount with real fee", fnName)
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Error().Err(err).Any("req", reqs).Msgf("%s - failed to commit transaction", fnName)
		return err
	}

	return nil
}

func (r *reportRepo) GetRealFee(ctx context.Context, tx *sqlx.Tx, registrationID string) (decimal.Decimal, error) {
	query := `
		SELECT
			COALESCE(pr.night_learning_fee, 0) +
			COALESCE(pr.foreign_learning_fee, 0) +
			COALESCE(pr.initial_fee,
				CASE
					WHEN pr.is_full_fee THEN (CASE WHEN pr.is_itp THEN pr.full_fee * 2 ELSE pr.full_fee END)
					ELSE (CASE WHEN pr.is_itp THEN pr.program_fee_per_meeting * 2 ELSE pr.program_fee_per_meeting END) * pr.program_meetings
				END
			) AS real_fee
		FROM
			program_registrations pr
		WHERE
			pr.id = ?
			AND pr.deleted_at IS NULL
	`
	var realFee decimal.Decimal
	err := tx.GetContext(ctx, &realFee, tx.Rebind(query), registrationID)
	if err != nil {
		return realFee, err
	}

	return realFee, nil
}

func (r *reportRepo) UpdateUsedAmountWithRealFee(ctx context.Context, tx *sqlx.Tx, registrationID string, realFee decimal.Decimal) error {
	query := `
		UPDATE program_registrations
		SET
			mentor_detail_fee_used = ?,
			notes_for_fund_distributions = NULL
		WHERE
			id = ?
			AND deleted_at IS NULL
			AND is_paid = TRUE
	`

	_, err := tx.ExecContext(ctx, tx.Rebind(query), realFee, registrationID)
	return err
}

func (r *reportRepo) GetAcquisitionRightsAggregate(ctx context.Context, req *entity.GetAcquisitionRightsAggregateReq) (
	*entity.GetAcquisitionRightsAggregateResp, error) {
	switch req.For {
	case "academic_manager":
		return r.GetAcquisitionRightsAggregateAcademicManager(ctx, req)
	case "student_manager":
		return r.GetAcquisitionRightsAggregateStudentManager(ctx, req)
	}

	return nil, fmt.Errorf("invalid for field")
}

func (r *reportRepo) GetAcquisitionRightsAggregateStudentManager(
	ctx context.Context,
	req *entity.GetAcquisitionRightsAggregateReq,
) (*entity.GetAcquisitionRightsAggregateResp, error) {
	fnName := "repo::GetAcquisitionRightsAggregateStudentManager"
	type dao struct {
		TotalData int `db:"total_data"`
		entity.AcquisitionRightsAggregate
	}
	var (
		resp = new(entity.GetAcquisitionRightsAggregateResp)
		data = make([]dao, 0)
		args = make([]any, 0, 3)
	)
	resp.Items = make([]entity.AcquisitionRightsAggregate, 0)

	query := `
		SELECT
			COUNT (*) OVER() AS total_data,
			TO_CHAR(pr.allocated_at AT TIME ZONE ?, 'YYYY-MM') AS month,
			sm.id AS student_manager_id,
			sm.name AS student_manager_name,
			SUM(
				CASE
					WHEN pr.program_acquisition_rights > 10 THEN 0
					WHEN (pr.is_paid = FALSE OR pr.mentor_detail_fee_used IS NULL OR pr.mentor_detail_fee_used = 0) THEN 0
					WHEN pr.is_itp THEN pr.program_acquisition_rights * 2
					ELSE pr.program_acquisition_rights
				END
			) AS total_acquisition_rights
		FROM
			program_registrations pr
		JOIN
			marketers m ON pr.marketer_id = m.id
		JOIN
			student_managers sm ON m.student_manager_id = sm.id
		WHERE
			pr.deleted_at IS NULL
	`
	args = append(args, req.Timezone)

	if req.Month != "" {
		query += ` AND TO_CHAR(pr.allocated_at AT TIME ZONE ?, 'YYYY-MM') = ?`
		args = append(args, req.Timezone, req.Month)
	}

	if req.StudentManagerId != "" {
		query += ` AND sm.id = ?`
		args = append(args, req.StudentManagerId)
	}

	query += `
		GROUP BY
			month,
			sm.id,
			sm.name
		ORDER BY
			sm.id ASC,
			month ASC
		LIMIT ? OFFSET ?
	`

	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	if err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...); err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to query acquisition rights", fnName)
		return nil, err
	}

	for _, d := range data {
		resp.Items = append(resp.Items, d.AcquisitionRightsAggregate)
		resp.Meta.TotalData = d.TotalData
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	return resp, nil
}

func (r *reportRepo) GetAcquisitionRightsAggregateAcademicManager(
	ctx context.Context,
	req *entity.GetAcquisitionRightsAggregateReq,
) (*entity.GetAcquisitionRightsAggregateResp, error) {
	fnName := "repo::GetAcquisitionRightsAggregateAcademicManager"
	type dao struct {
		TotalData int `db:"total_data"`
		entity.AcquisitionRightsAggregate
	}
	var (
		resp = new(entity.GetAcquisitionRightsAggregateResp)
		data = make([]dao, 0)
		args = make([]any, 0, 3)
	)

	resp.Items = make([]entity.AcquisitionRightsAggregate, 0)

	query := `
		SELECT
			COUNT (*) OVER() AS total_data,
			TO_CHAR(pr.allocated_at AT TIME ZONE ?, 'YYYY-MM') AS month,
			am.id AS academic_manager_id,
			am.name AS academic_manager_name,
			SUM(
				CASE
					WHEN pr.program_acquisition_rights > 10 THEN 0
					WHEN (pr.is_paid = FALSE OR pr.mentor_detail_fee_used IS NULL OR pr.mentor_detail_fee_used = 0) THEN 0
					WHEN pr.is_itp THEN pr.program_acquisition_rights * 2
					ELSE pr.program_acquisition_rights
				END
			) AS total_acquisition_rights
		FROM
			program_registrations pr
		JOIN
			lecturers l ON pr.lecturer_id = l.id
		JOIN
			academic_managers am ON l.academic_manager_id = am.id
		WHERE
			pr.deleted_at IS NULL
	`
	args = append(args, req.Timezone)

	if req.Month != "" {
		query += ` AND TO_CHAR(pr.allocated_at AT TIME ZONE ?, 'YYYY-MM') = ?`
		args = append(args, req.Timezone, req.Month)
	}

	if req.AcademicManagerID != "" {
		query += ` AND am.id = ?`
		args = append(args, req.AcademicManagerID)
	}

	query += `
		GROUP BY
			month,
			am.id,
			am.name
		ORDER BY
			am.id ASC,
			month ASC
		LIMIT ? OFFSET ?
	`

	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	if err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...); err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to query acquisition rights", fnName)
		return nil, err
	}

	for _, d := range data {
		resp.Items = append(resp.Items, d.AcquisitionRightsAggregate)
		resp.Meta.TotalData = d.TotalData
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	return resp, nil
}
