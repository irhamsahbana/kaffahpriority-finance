package repository

import (
	"codebase-app/internal/module/report/entity"
	"codebase-app/pkg/errmsg"
	"context"
	"fmt"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (r *reportRepo) CreateRegistrations(ctx context.Context, req *entity.CreateRegistrationsReq) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("repo::CreateRegistrations - failed to begin transaction")
		return err
	}
	defer func() {
		if err != nil {
			errRB := tx.Rollback()
			if errRB != nil {
				log.Error().Err(errRB).Any("req", req).Msg("repo::CreateRegistrations - failed to rollback transaction")
			}
			return
		}
		errCommit := tx.Commit()
		if errCommit != nil {
			log.Error().Err(errCommit).Any("req", req).Msg("repo::CreateRegistrations - failed to commit transaction")
		}
	}()

	query := `
		WITH template AS (
			SELECT
				prt.id,
				prt.program_id,
				prt.lecturer_id,
				prt.marketer_id,
				prt.student_id,
				prt.is_itp,
				p.name AS program_name,
				p.acquisition_rights,
				p.full_fee,
				prt.program_fee,
				prt.program_fee_per_meeting,
				prt.administration_fee,
				prt.foreign_learning_fee,
				prt.night_learning_fee,
				prt.marketer_commission_fee,
				prt.overpayment_fee,
				prt.hr_fee,
				prt.marketer_gifts_fee,
				prt.closing_fee_for_office,
				prt.closing_fee_for_reward,
				prt.days,
				prt.notes
			FROM
				program_registration_templates prt
			JOIN
				programs p
				ON prt.program_id = p.id
			WHERE
				prt.id = ?
				AND prt.deleted_at IS NULL
		)
		INSERT INTO program_registrations (
		id,
		template_id,
		user_id,
		program_id,
		lecturer_id,
		marketer_id,
		student_id,
		program_name,
		program_fee,
		program_fee_per_meeting,
		full_fee,
		program_meetings,
		program_acquisition_rights,
		administration_fee,
		foreign_learning_fee,
		night_learning_fee,
		is_itp,
		marketer_commission_fee,
		overpayment_fee,
		hr_fee,
		mentor_detail_fee,
		mentor_detail_fee_used,
		hr_detail_fee,
		marketer_gifts_fee,
		closing_fee_for_office,
		closing_fee_for_reward,
		days,
		notes,
		allocated_at
		)
		SELECT
			?,
			(SELECT id FROM template),
			?,
			(SELECT program_id FROM template),
			(SELECT lecturer_id FROM template),
			(SELECT marketer_id FROM template),
			(SELECT student_id FROM template),
			(SELECT program_name FROM template),
			(SELECT program_fee FROM template),
			(SELECT program_fee_per_meeting FROM template),
			(SELECT full_fee FROM template),
			0,
			(SELECT acquisition_rights FROM template),
			(SELECT administration_fee FROM template),
			(SELECT foreign_learning_fee FROM template),
			(SELECT night_learning_fee FROM template),
			(SELECT is_itp FROM template),
			(SELECT marketer_commission_fee FROM template),
			(SELECT overpayment_fee FROM template),
			(SELECT hr_fee FROM template),
			(SELECT hr_fee - 40000 FROM template),
			(SELECT hr_fee - 40000 FROM template),
			40000,
			(SELECT marketer_gifts_fee FROM template),
			(SELECT closing_fee_for_office FROM template),
			(SELECT closing_fee_for_reward FROM template),
			(SELECT days FROM template),
			(SELECT notes FROM template),
			NOW()
		`

	queryStudents := `
		SELECT
			adds.student_id,
			adds.name
		FROM
			prt_additional_students adds
		WHERE
			adds.prt_id = ?
	`
	queryStudents = r.db.Rebind(queryStudents)

	queryInsertStudents := `
		INSERT INTO pr_additional_students (
			id,
			pr_id,
			student_id,
			name
		) VALUES (?, ?, ?, ?)
	`
	queryInsertStudents = r.db.Rebind(queryInsertStudents)

	for _, item := range req.Registrations {
		var prId = ulid.Make().String()
		var students = make([]entity.AddStudent, 0)

		// check if program_id, lecturer_id, and student_id already exist in this month

		queryCheck := `
			SELECT EXISTS (
				WITH template AS (
					SELECT
						prt.program_id,
						prt.lecturer_id,
						prt.student_id
					FROM
						program_registration_templates prt
					WHERE
						prt.id = ?
						AND prt.deleted_at IS NULL
				)
				SELECT
					1
				FROM
					program_registrations pr
				WHERE
					pr.program_id = (SELECT program_id FROM template)
					-- AND pr.lecturer_id = (SELECT lecturer_id FROM template)
					AND (
						CASE
							WHEN pr.lecturer_id IS NULL THEN (SELECT lecturer_id FROM template) IS NULL
							ELSE pr.lecturer_id = (SELECT lecturer_id FROM template)
						END
					)
					AND pr.student_id = (SELECT student_id FROM template)
					AND EXTRACT(MONTH FROM pr.allocated_at) = EXTRACT(MONTH FROM NOW())
					AND EXTRACT(YEAR FROM pr.allocated_at) = EXTRACT(YEAR FROM NOW())
					AND pr.deleted_at IS NULL
			)
		`

		var exist bool
		err = tx.GetContext(ctx, &exist, tx.Rebind(queryCheck), item.TemplateId)
		if err != nil {
			log.Error().Err(err).Any("req", req).Any("template_id", item.TemplateId).Msg("repo::CreateRegistrations - failed to check data")
			return err
		}

		if exist {
			log.Warn().Any("req", req).Any("template_id", item.TemplateId).Msg("repo::CreateRegistrations - data already exist")
			err = errmsg.NewCustomErrors(403).SetMessage(`Data dengan template id ` + item.TemplateId + ` sudah dibuat di bulan ini`)
			return err
		}

		// check if the template's fields (lecturer_id, marketer_id) are null
		// if so, return error with message "Template perlu dilengkapi (pengajar, marketer)"
		queryCheckTemplate := `
			SELECT EXISTS (
				SELECT
					1
				FROM
					program_registration_templates prt
				WHERE
					prt.id = ?
					AND prt.deleted_at IS NULL
					AND prt.lecturer_id IS NOT NULL
					AND prt.marketer_id IS NOT NULL
			)
		`
		queryCheckTemplate = r.db.Rebind(queryCheckTemplate)
		var isTemplateValid bool
		err = tx.GetContext(ctx, &isTemplateValid, queryCheckTemplate, item.TemplateId)
		if err != nil {
			log.Error().Err(err).Any("req", req).Any("template_id", item.TemplateId).Msg("repo::CreateRegistrations - failed to check template")
			return err
		}
		if !isTemplateValid {
			log.Warn().Any("req", req).Any("template_id", item.TemplateId).Msg("repo::CreateRegistrations - template is not valid")
			err = errmsg.NewCustomErrors(403).SetMessage(`Template perlu dilengkapi (pengajar, marketer)`)
			return err
		}

		_, err = tx.ExecContext(ctx, tx.Rebind(query),
			item.TemplateId,
			prId, req.UserId,
			// item.IsFirstRegistration,
		)
		if err != nil {
			log.Error().Err(err).Any("req", req).Any("template_id", item.TemplateId).Msg("repo::CreateRegistrations - failed to insert data")
			return err
		}

		// fetch additional students from prt_additional_students
		err = tx.SelectContext(ctx, &students, queryStudents, item.TemplateId)
		if err != nil {
			log.Error().Err(err).Any("req", req).Any("template_id", item.TemplateId).Msg("repo::CreateRegistrations - failed to fetch additional students")
			return err
		}

		// insert additional students into pr_additional_students
		for _, student := range students {
			_, err = tx.ExecContext(ctx, queryInsertStudents,
				ulid.Make().String(), prId, student.StudentId, student.Name,
			)
			if err != nil {
				log.Error().Err(err).Any("req", req).Any("template_id", item.TemplateId).Msg("repo::CreateRegistrations - failed to insert additional students")
				return err
			}
		}

	}

	return nil
}

func (r *reportRepo) CopyRegistrations(ctx context.Context, req *entity.CopyRegistrationsReq) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("repo::CopyRegistrations - failed to begin transaction")
		return err
	}
	defer func() {
		if err != nil {
			errRB := tx.Rollback()
			if errRB != nil {
				log.Error().Err(errRB).Any("req", req).Msg("repo::CopyRegistrations - failed to rollback transaction")
			}
			return
		}
		errCommit := tx.Commit()
		if errCommit != nil {
			log.Error().Err(errCommit).Any("req", req).Msg("repo::CopyRegistrations - failed to commit transaction")
		}
	}()

	queryCheck := `
			SELECT EXISTS (
				WITH regis AS (
					SELECT
						r.program_id,
						r.lecturer_id,
						r.student_id
					FROM
						program_registrations r
					WHERE
						r.id = ?
						AND r.deleted_at IS NULL
				)
				SELECT
					1
				FROM
					program_registrations pr
				WHERE
					pr.program_id = (SELECT program_id FROM regis)
					AND (
						CASE
							WHEN pr.lecturer_id IS NULL THEN (SELECT lecturer_id FROM regis) IS NULL
							ELSE pr.lecturer_id = (SELECT lecturer_id FROM regis)
						END
					)
					AND pr.student_id = (SELECT student_id FROM regis)
					AND EXTRACT(MONTH FROM pr.allocated_at) = EXTRACT(MONTH FROM ?::timestamptz AT TIME ZONE ?)
					AND EXTRACT(YEAR FROM pr.allocated_at) = EXTRACT(YEAR FROM ?::timestamptz AT TIME ZONE ?)
					AND pr.deleted_at IS NULL
			)
		`
	queryCheck = r.db.Rebind(queryCheck)

	query := `
		INSERT INTO program_registrations (
		id,
		template_id,
		user_id,
		program_id,
		lecturer_id,
		marketer_id,
		student_id,
		program_name,
		program_fee,
		program_meetings,
		program_acquisition_rights,
		foreign_learning_fee,
		night_learning_fee,
		marketer_commission_fee,
		overpayment_fee,
		hr_fee,
		mentor_detail_fee,
		mentor_detail_fee_used,
		hr_detail_fee,
		days,
		notes,
		allocated_at
		)
		SELECT
			?,
			pr.template_id,
			?,
			pr.program_id,
			pr.lecturer_id,
			pr.marketer_id,
			pr.student_id,
			pr.program_name,
			pr.program_fee,
			pr.program_meetings,
			pr.program_acquisition_rights,
			pr.foreign_learning_fee,
			pr.night_learning_fee,
			pr.marketer_commission_fee,
			pr.overpayment_fee,
			pr.hr_fee,
			pr.mentor_detail_fee,
			pr.mentor_detail_fee_used,
			pr.hr_detail_fee,
			pr.days,
			pr.notes,
			?::timestamptz AT TIME ZONE ?
		FROM
			program_registrations pr
		WHERE
			pr.id = ?
			AND pr.deleted_at IS NULL
	`

	queryStudents := `
		SELECT
			adds.student_id,
			adds.name
		FROM
			pr_additional_students adds
		WHERE
			adds.pr_id = ?
	`
	queryStudents = r.db.Rebind(queryStudents)

	queryInsertStudents := `
		INSERT INTO pr_additional_students (
			id,
			pr_id,
			student_id,
			name
		) VALUES (?, ?, ?, ?)
	`
	queryInsertStudents = r.db.Rebind(queryInsertStudents)

	for _, item := range req.Registrations {
		var prId = ulid.Make().String()
		var students = make([]entity.AddStudent, 0)

		// check if registration_id already exist in this month
		var exist bool
		err = tx.GetContext(ctx, &exist, queryCheck, item.RegisId,
			item.AllocatedAt, item.Timezone,
			item.AllocatedAt, item.Timezone,
		)
		if err != nil {
			log.Error().Err(err).Any("req", req).Any("registration_id", item.RegisId).Msg("repo::CopyRegistrations - failed to check data")
		}

		if exist {
			log.Warn().Any("req", req).Any("registration_id", item.RegisId).Msg("repo::CopyRegistrations - data already exist")
			err = errmsg.NewCustomErrors(403).SetMessage(fmt.Sprintf(`Data yang sama (program, pengajar dan murid) sudah dialokasikan pada %s (timezone %s)`, item.AllocatedAt, item.Timezone))
			return err
		}

		// create new registration based on existing registration
		_, err = tx.ExecContext(ctx, tx.Rebind(query),
			prId, req.UserId,
			item.AllocatedAt, item.Timezone,
			item.RegisId,
		)
		if err != nil {
			log.Error().Err(err).Any("req", req).Any("registration_id", item.RegisId).Msg("repo::CopyRegistrations - failed to insert data")
			return err
		}

		err = tx.SelectContext(ctx, &students, queryStudents, item.RegisId)
		if err != nil {
			log.Error().Err(err).Any("req", req).Any("registration_id", item.RegisId).Msg("repo::CopyRegistrations - failed to fetch additional students")
			return err
		}

		for _, student := range students {
			_, err = tx.ExecContext(ctx, queryInsertStudents,
				ulid.Make().String(), prId, student.StudentId, student.Name,
			)
			if err != nil {
				log.Error().Err(err).Any("req", req).Any("registration_id", item.RegisId).Msg("repo::CopyRegistrations - failed to insert additional students")
				return err
			}
		}

	}

	return nil
}
