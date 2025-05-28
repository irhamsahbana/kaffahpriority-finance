package repository

import (
	"codebase-app/internal/module/report/entity"
	"codebase-app/pkg/errmsg"
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
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
			am.name AS academic_manager_name,
			l.name AS lecturer_name,
			m.name AS marketer_name,
			sm.name AS student_manager_name,
			pr.foreign_learning_fee,
			pr.night_learning_fee,
			pr.is_itp,
			CASE
				WHEN pr.is_itp THEN pr.program_acquisition_rights * 2
				ELSE pr.program_acquisition_rights
			END AS acquisition_rights,
			pr.program_meetings,
			pr.program_fee_per_meeting,
			COALESCE(pr.initial_fee, (
				CASE
					WHEN pr.initial_fee IS NOT NULL THEN pr.initial_fee
					WHEN pr.is_full_fee = TRUE THEN pr.full_fee
					ELSE pr.program_fee_per_meeting * pr.program_meetings
				END
				)
			) AS initial_fee,
			(
				COALESCE(pr.night_learning_fee, 0) +
				COALESCE(pr.foreign_learning_fee, 0) +
				COALESCE(pr.initial_fee,
					CASE
						WHEN pr.is_full_fee THEN pr.full_fee
						ELSE pr.program_fee_per_meeting * pr.program_meetings
					END
				)
			) AS real_fee,
			pr.is_full_fee,
			pr.full_fee,
			-- pr.mentor_detail_fee_used,
			CASE
				WHEN pr.is_paid = TRUE THEN pr.mentor_detail_fee_used
				ELSE 0
			END AS mentor_detail_fee_used,
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
		JOIN
			student_managers sm ON m.student_manager_id = sm.id
		WHERE
			pr.deleted_at IS NULL
			AND TO_CHAR(pr.allocated_at AT TIME ZONE ?, 'YYYY-MM') = ?
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

	if req.Q != "" {
		query += ` AND (pr.program_name ILIKE '%' || ? || '%' OR s.name ILIKE '%' || ? || '%' OR l.name ILIKE '%' || ? || '%')`
		args = append(args, req.Q, req.Q, req.Q)
	}

	query += `
		ORDER BY
			pr.lecturer_id ASC,
			pr.student_id ASC,
			pr.template_id ASC
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

func (r *reportRepo) GetLecturersWagesAggregate(ctx context.Context, req *entity.GetLecturersWagesAggregateReq) (*entity.LecturersWageAggregateResp, error) {
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
								WHEN pr.is_full_fee THEN pr.full_fee
								ELSE pr.program_fee_per_meeting * pr.program_meetings
							END
						)
					)
				),
				0
			) AS total_real_fee,
			COALESCE(
				SUM(
					CASE
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

	if req.AcademicManagerId != "" {
		query += ` AND am.id = ?`
		args = append(args, req.AcademicManagerId)
	}

	if req.LecturerId != "" {
		query += ` AND pr.lecturer_id = ?`
		args = append(args, req.LecturerId)
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
		log.Error().Err(err).Any("req", req).Msg("repo::GetLecturersAggregateWages - failed to query lecturers wages")
		return nil, err
	}

	for _, d := range data {
		resp.Items = append(resp.Items, d.LecturersWageAggregateItem)
		resp.Meta.TotalData = d.TotalData
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	return resp, nil
}

func (r *reportRepo) UpdateLecturersWage(ctx context.Context, req *entity.UpdateLecturersWageReq) error {
	queryParts := []string{}
	args := []any{}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateLecturersWages - failed to begin transaction")
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
	args = append(args, req.RegistrationId)

	_, err = r.db.ExecContext(ctx, r.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateLecturersWages - failed to update lecturers wages")
		return err
	}

	// jika ada perubahan, update juga used_amount menggunakan ujroh real
	// 1. cari kalkulasi real fee

	query = `
		SELECT
			COALESCE(pr.night_learning_fee, 0) +
			COALESCE(pr.foreign_learning_fee, 0) +
			COALESCE(pr.initial_fee,
				CASE
					WHEN pr.is_full_fee THEN pr.full_fee
					ELSE pr.program_fee_per_meeting * pr.program_meetings
				END
			) AS real_fee
		FROM
			program_registrations pr
		WHERE
			pr.id = ?
			AND pr.deleted_at IS NULL
	`
	var realFee decimal.Decimal
	err = tx.GetContext(ctx, &realFee, tx.Rebind(query), req.RegistrationId)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warn().Err(err).Any("req", req).Msg("repo::UpdateLecturersWages - real fee not found")
			return errmsg.NewCustomErrors(404).SetMessage("Laporan tidak ditemukan")
		}
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateLecturersWages - failed to get real fee")
		return err
	}

	// 2. update used_amount menggunakan real fee
	query = `
		UPDATE program_registrations
		SET
			mentor_detail_fee_used = ?,
			notes_for_fund_distributions = NULL
		WHERE
			id = ?
			AND deleted_at IS NULL
			AND is_paid = TRUE
	`

	_, err = tx.ExecContext(ctx, tx.Rebind(query), realFee, req.RegistrationId)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateLecturersWages - failed to update ujroh real")
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateLecturersWages - failed to commit transaction")
		return err
	}

	return nil
}

func (r *reportRepo) GetAcquisitionRightsAggregate(ctx context.Context, req *entity.GetAcquisitionRightsAggregateReq) (
	*entity.GetAcquisitionRightsAggregateResp, error) {
	if req.For == "academic_manager" {
		return r.GetAcquisitionRightsAggregateAcademicManager(ctx, req)
	} else if req.For == "student_manager" {
		return r.GetAcquisitionRightsAggregateStudentManager(ctx, req)
	}

	return nil, fmt.Errorf("invalid for field")
}

func (r *reportRepo) GetAcquisitionRightsAggregateStudentManager(
	ctx context.Context,
	req *entity.GetAcquisitionRightsAggregateReq,
) (*entity.GetAcquisitionRightsAggregateResp, error) {
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
		log.Error().Err(err).Any("req", req).Msg("repo::GetAcquisitionRightsAggregateStudentManager - failed to query acquisition rights")
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

	if req.AcademicManagerId != "" {
		query += ` AND am.id = ?`
		args = append(args, req.AcademicManagerId)
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
		log.Error().Err(err).Any("req", req).Msg("repo::GetAcquisitionRightsAggregate - failed to query acquisition rights")
		return nil, err
	}

	for _, d := range data {
		resp.Items = append(resp.Items, d.AcquisitionRightsAggregate)
		resp.Meta.TotalData = d.TotalData
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	return resp, nil
}
