package repository

import (
	"codebase-app/internal/module/report/entity"
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (r *reportRepo) GenerateRegistrationReports(ctx context.Context, req *entity.GenerateRegistrationsReq) error {
	fnName := "repo::GenerateRegistrationReports"
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msgf("%s - failed to begin transaction", fnName)
		return err
	}
	defer tx.Rollback()

	// get all templates from the database that are not deleted
	templateIds := []string{}

	queryTemplates := `
		SELECT
			id
		FROM
			program_registration_templates
		WHERE
			deleted_at IS NULL
			AND marketer_id IS NOT NULL
		ORDER BY
			created_at ASC
	`

	// query all templates that are not deleted and have a marketer_id set
	if err := tx.SelectContext(ctx, &templateIds, queryTemplates); err != nil {
		log.Error().Err(err).Msgf("%s - failed to select templates", fnName)
		return err
	}

	// query all deleted templates
	deletedTemplateIds := []string{}
	queryDeletedTemplates := `
		SELECT
			id
		FROM
			program_registration_templates
		WHERE
			deleted_at IS NOT NULL
	`

	if err := tx.SelectContext(ctx, &deletedTemplateIds, queryDeletedTemplates); err != nil {
		log.Error().Err(err).Msgf("%s - failed to select deleted templates", fnName)
		return err
	}

	// delete registration that their master (template) has been deleted
	if len(deletedTemplateIds) > 0 {
		queryDeleteRegistrations := `
			UPDATE
				program_registrations
			SET
				deleted_at = NOW()
			WHERE
				template_id = ANY($1)
				AND deleted_at IS NULL
				AND is_paid = FALSE
				AND TO_CHAR(created_at AT TIME ZONE $2, 'YYYY-MM') = TO_CHAR(NOW() AT TIME ZONE $2, 'YYYY-MM')
		`

		if _, err := tx.ExecContext(ctx, queryDeleteRegistrations,
			pq.Array(deletedTemplateIds),
			req.Timezone,
		); err != nil {
			log.Error().Err(err).Msgf("%s - failed to delete registrations", fnName)
			return err
		}
	}

	// setup create registration batch
	batchId := ulid.Make().String()
	queryInsertRegistration := r.db.Rebind(queryInsertRegistration)
	queryStudents := r.db.Rebind(queryStudents)
	queryInsertStudents := r.db.Rebind(queryInsertStudents)

	for _, templateId := range templateIds {
		// Ambil data dari template untuk pengecekan
		var studentId, programId string
		var lecturerId *string
		err = tx.QueryRowContext(ctx, `SELECT lecturer_id, student_id, program_id FROM program_registration_templates WHERE id = $1`,
			templateId).Scan(&lecturerId, &studentId, &programId)
		if err != nil {
			log.Error().Err(err).Str("templateId", templateId).Any("req", req).Msgf("%s - failed to get template data", fnName)
			return err
		}

		// check if registration already exists
		var programName, studentName string
		var lecturerName *string
		err = tx.QueryRowxContext(ctx, r.db.Rebind(queryCheckRegistrationExists),
			lecturerId, lecturerId,
			studentId, programId, req.Timezone, req.Timezone, req.Timezone, req.Timezone,
		).Scan(&programName, &lecturerName, &studentName)

		// if registration already exists, skip this template
		if err == nil {
			continue
		} else if err != sql.ErrNoRows {
			log.Error().Err(err).Str("templateId", templateId).Any("req", req).Msgf("%s - failed to check if registration exists", fnName)
			return err
		}

		// if registration does not exist, proceed to insert

		// insert into program_registrations
		programRegistrationId := ulid.Make().String()
		var students = make([]entity.AddStudent, 0)

		args := []any{
			templateId,
			programRegistrationId,
			req.UserID,
			batchId,
		}

		// insert into program_registrations
		if _, err := tx.ExecContext(ctx, queryInsertRegistration, args...); err != nil {
			log.Error().Err(err).Any("req", req).Msgf("%s - failed to insert program registration", fnName)
			return err
		}

		// fetch additional students from prt_additional_students
		err = tx.SelectContext(ctx, &students, queryStudents, templateId)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msgf("%s - failed to select additional students", fnName)
			return err
		}

		// insert into pr_additional_students
		for _, student := range students {
			_, err = tx.ExecContext(ctx, queryInsertStudents,
				ulid.Make().String(), programRegistrationId, student.StudentID, student.Name,
			)
			if err != nil {
				log.Error().Err(err).Any("req", req).Any("template_id", templateId).Msgf("%s - failed to insert additional students", fnName)
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msgf("%s - failed to commit transaction", fnName)
		return err
	}

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
	NULL,
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

var queryCheckRegistrationExists = `
	SELECT
		p.name AS program_name,
		l.name AS lecturer_name,
		s.name AS student_name
	FROM
		program_registrations pr
	JOIN
		programs p
		ON pr.program_id = p.id
	LEFT JOIN
		lecturers l
		ON pr.lecturer_id = l.id
	JOIN
		students s
		ON pr.student_id = s.id
	WHERE
		pr.deleted_at IS NULL
		AND (
			(pr.lecturer_id IS NULL AND ?::TEXT IS NULL)
			OR
			(pr.lecturer_id = ?)
		)
		AND pr.student_id = ?
		AND pr.program_id = ?
		AND EXTRACT(YEAR FROM pr.allocated_at AT TIME ZONE ?) = EXTRACT(YEAR FROM NOW() AT TIME ZONE ?)
		AND EXTRACT(MONTH FROM pr.allocated_at AT TIME ZONE ?) = EXTRACT(MONTH FROM NOW() AT TIME ZONE ?)
	LIMIT 1
`
