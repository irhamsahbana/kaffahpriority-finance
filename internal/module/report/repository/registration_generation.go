package repository

import (
	"codebase-app/internal/module/report/entity"
	"context"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (r *reportRepo) GenerateRegistrationReports(ctx context.Context, req *entity.GenerateRegistrationsReq) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("repo::GenerateRegistrationReport - failed to begin transaction")
	}
	defer tx.Rollback()

	batchId := ulid.Make().String()

	// get all templates from the database
	templateIds := []string{}
	queryTemplates := `
		SELECT
			id
		FROM
			program_registration_templates
		WHERE
			deleted_at IS NULL
			AND marketer_id IS NOT NULL
			AND lecturer_id IS NOT NULL
	`

	// query all templates that are not deleted and have a marketer_id and lecturer_id
	if err := tx.SelectContext(ctx, &templateIds, queryTemplates); err != nil {
		log.Error().Err(err).Msg("repo::GenerateRegistrationReport - failed to select templates")
		return err
	}

	log.Debug().Any("template_ids", templateIds).Msg("repo::GenerateRegistrationReport - selected templates")

	queryStudents := r.db.Rebind(queryStudents)
	queryInsertStudents := r.db.Rebind(queryInsertStudents)

	for _, templateId := range templateIds {
		// insert into program_registrations
		programRegistrationId := ulid.Make().String()
		var students = make([]entity.AddStudent, 0)

		queryInsertRegistration := r.db.Rebind(queryInsertRegistration)
		args := []any{
			templateId,
			programRegistrationId,
			req.UserId,
			batchId,
		}

		log.Debug().Any("args", args).Msg("repo::GenerateRegistrationReport - inserting program registration")

		// insert into program_registrations
		if _, err := tx.ExecContext(ctx, queryInsertRegistration, args...); err != nil {
			log.Error().Err(err).Msg("repo::GenerateRegistrationReport - failed to insert program registration")
			return err
		}

		// fetch additional students from prt_additional_students
		err = tx.SelectContext(ctx, &students, queryStudents, templateId)
		if err != nil {
			log.Error().Err(err).Msg("repo::GenerateRegistrationReport - failed to select students")
			return err
		}

		// insert into pr_additional_students
		for _, student := range students {
			_, err = tx.ExecContext(ctx, queryInsertStudents,
				ulid.Make().String(), programRegistrationId, student.StudentId, student.Name,
			)
			if err != nil {
				log.Error().Err(err).Any("req", req).Any("template_id", templateId).Msg("repo::GenerateRegistrationReport - failed to insert additional students")
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("repo::GenerateRegistrationReport - failed to commit transaction")
		return err
	}

	log.Debug().Msg("repo::GenerateRegistrationReport - successfully generated registration reports")

	return nil
}

var queryInsertRegistration = `
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
		batch,
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
		started_at
		)
		SELECT
			?,
			(SELECT id FROM template),
			?,
			(SELECT program_id FROM template),
			(SELECT lecturer_id FROM template),
			(SELECT marketer_id FROM template),
			(SELECT student_id FROM template),
			?,
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

var queryStudents = `
		SELECT
			adds.student_id,
			adds.name
		FROM
			prt_additional_students adds
		WHERE
			adds.prt_id = ?
	`

var queryInsertStudents = `
		INSERT INTO pr_additional_students (
			id,
			pr_id,
			student_id,
			name
		) VALUES (?, ?, ?, ?)
	`
