package repository

import (
	"codebase-app/internal/module/report/entity"
	"codebase-app/pkg/errmsg"
	"context"

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
		administration_fee,
		foreign_learning_fee,
		night_learning_fee,
		marketer_commission_fee,
		overpayment_fee,
		hr_fee,
		marketer_gifts_fee,
		closing_fee_for_office,
		closing_fee_for_reward,
		days,
		notes,
		started_at
		)
		SELECT
			?,
			?,
			?,
			prt.program_id,
			prt.lecturer_id,
			prt.marketer_id,
			prt.student_id,
			p.name,
			prt.program_fee,
			0,
			CASE
				WHEN ? = TRUE THEN prt.administration_fee
				ELSE NULL
			END,
			prt.foreign_learning_fee,
			prt.night_learning_fee,
			prt.marketer_commission_fee,
			prt.overpayment_fee,
			prt.hr_fee,
			prt.marketer_gifts_fee,
			prt.closing_fee_for_office,
			prt.closing_fee_for_reward,
			prt.days,
			prt.notes,
			NOW()
		FROM
			program_registration_templates prt
		JOIN
			programs p
			ON prt.program_id = p.id
		WHERE
			prt.id = ?
			AND prt.deleted_at IS NULL
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
					AND pr.lecturer_id = (SELECT lecturer_id FROM template)
					AND pr.student_id = (SELECT student_id FROM template)
					AND EXTRACT(MONTH FROM pr.started_at) = EXTRACT(MONTH FROM NOW())
					AND EXTRACT(YEAR FROM pr.started_at) = EXTRACT(YEAR FROM NOW())
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
			err = errmsg.NewCustomErrors(403).SetMessage(`Data dengan template id ` + item.TemplateId + ` sudah ada di bulan ini`)
			return err
		}

		_, err = tx.ExecContext(ctx, tx.Rebind(query),
			prId, item.TemplateId, req.UserId, item.IsFirstRegistration, item.TemplateId,
		)
		if err != nil {
			log.Error().Err(err).Any("req", req).Any("template_id", item.TemplateId).Msg("repo::CreateRegistrations - failed to insert data")
			return err
		}

		err = tx.SelectContext(ctx, &students, queryStudents, item.TemplateId)
		if err != nil {
			log.Error().Err(err).Any("req", req).Any("template_id", item.TemplateId).Msg("repo::CreateRegistrations - failed to fetch additional students")
			return err
		}

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
		administration_fee,
		foreign_learning_fee,
		night_learning_fee,
		marketer_commission_fee,
		overpayment_fee,
		hr_fee,
		marketer_gifts_fee,
		closing_fee_for_office,
		closing_fee_for_reward,
		days,
		notes
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
			pr.administration_fee,
			pr.foreign_learning_fee,
			pr.night_learning_fee,
			pr.marketer_commission_fee,
			pr.overpayment_fee,
			pr.hr_fee,
			pr.marketer_gifts_fee,
			pr.closing_fee_for_office,
			pr.closing_fee_for_reward,
			pr.days,
			pr.notes
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

		_, err = tx.ExecContext(ctx, tx.Rebind(query),
			prId, req.UserId, item.RegisId,
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
