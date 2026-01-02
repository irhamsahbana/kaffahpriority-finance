package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/pkg/errmsg"
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (r *reportRepo) CreateRegistrations(ctx context.Context, req *entity.CreateRegistrationsReq) error {
	ctx, span := tracing.StartSpan(ctx, "repo.CreateRegistrations")
	defer span.End()

	var err error

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("failed to begin transaction")
		return err
	}
	defer func() {
		if err != nil {
			if errRB := tx.Rollback(); errRB != nil {
				log.Ctx(ctx).Error().Err(errRB).Any("req", req).Msgf("failed to rollback transaction")
			}
			return
		}
		if errCommit := tx.Commit(); errCommit != nil {
			log.Ctx(ctx).Error().Err(errCommit).Any("req", req).Msgf("failed to commit transaction")
		}
	}()

	for _, item := range req.Registrations {
		err = r.processRegistration(ctx, tx, req.UserID, item)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *reportRepo) processRegistration(ctx context.Context, tx *sqlx.Tx, userID string, item entity.RegistrationItem) error {
	ctx, span := tracing.StartSpan(ctx, "repo.processRegistration")
	defer span.End()

	if err := r.checkRegistrationExists(ctx, tx, item.TemplateID); err != nil {
		return err
	}

	if err := r.validateTemplate(ctx, tx, item.TemplateID); err != nil {
		return err
	}

	prID := ulid.Make().String()
	if err := r.insertRegistration(ctx, tx, userID, item.TemplateID, prID); err != nil {
		return err
	}

	if err := r.copyAdditionalStudents(ctx, tx, item.TemplateID, prID); err != nil {
		return err
	}

	return nil
}

func (r *reportRepo) checkRegistrationExists(ctx context.Context, tx *sqlx.Tx, templateID string) error {
	ctx, span := tracing.StartSpan(ctx, "repo.checkRegistrationExists")
	defer span.End()

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
	if err := tx.GetContext(ctx, &exist, tx.Rebind(queryCheck), templateID); err != nil {
		log.Ctx(ctx).Error().Err(err).Str("template_id", templateID).Msg("failed to check existing registration")
		return err
	}

	if exist {
		log.Ctx(ctx).Warn().Str("template_id", templateID).Msg("registration already exists")
		return errmsg.NewCustomErrors(403).SetMessage(`Data dengan template id ` + templateID + ` sudah dibuat di bulan ini`)
	}

	return nil
}

func (r *reportRepo) validateTemplate(ctx context.Context, tx *sqlx.Tx, templateID string) error {
	ctx, span := tracing.StartSpan(ctx, "repo.validateTemplate")
	defer span.End()

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

	var isTemplateValid bool
	if err := tx.GetContext(ctx, &isTemplateValid, tx.Rebind(queryCheckTemplate), templateID); err != nil {
		log.Ctx(ctx).Error().Err(err).Str("template_id", templateID).Msg("failed to validate template")
		return err
	}

	if !isTemplateValid {
		log.Ctx(ctx).Warn().Str("template_id", templateID).Msg("template is incomplete")
		return errmsg.NewCustomErrors(403).SetMessage(`Template perlu dilengkapi (pengajar, marketer)`)
	}

	return nil
}

func (r *reportRepo) insertRegistration(ctx context.Context, tx *sqlx.Tx, userID, templateID, prID string) error {
	ctx, span := tracing.StartSpan(ctx, "repo.insertRegistration")
	defer span.End()

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
			(CASE WHEN (SELECT is_itp FROM template) THEN 2 ELSE 1 END * (SELECT acquisition_rights FROM template)),
			(SELECT administration_fee FROM template),
			(SELECT foreign_learning_fee FROM template),
			(SELECT night_learning_fee FROM template),
			(SELECT is_itp FROM template),
			(SELECT marketer_commission_fee FROM template),
			(SELECT overpayment_fee FROM template),
			(SELECT hr_fee FROM template),
			(SELECT hr_fee - (40000 * CASE WHEN is_itp THEN 2 ELSE 1 END * acquisition_rights) FROM template),
			null,
			(40000 * CASE WHEN (SELECT is_itp FROM template) THEN 2 ELSE 1 END * (SELECT acquisition_rights FROM template)),
			(SELECT marketer_gifts_fee FROM template),
			(SELECT closing_fee_for_office FROM template),
			(SELECT closing_fee_for_reward FROM template),
			(SELECT days FROM template),
			(SELECT notes FROM template),
			NOW()
	`

	_, err := tx.ExecContext(ctx, tx.Rebind(query), templateID, prID, userID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("template_id", templateID).Msg("failed to insert registration")
		return err
	}

	return nil
}

func (r *reportRepo) copyAdditionalStudents(ctx context.Context, tx *sqlx.Tx, templateID, prID string) error {
	ctx, span := tracing.StartSpan(ctx, "repo.copyAdditionalStudents")
	defer span.End()

	queryStudents := `
		SELECT
			adds.student_id,
			adds.name
		FROM
			prt_additional_students adds
		WHERE
			adds.prt_id = ?
	`

	var students []entity.AddStudent
	if err := tx.SelectContext(ctx, &students, tx.Rebind(queryStudents), templateID); err != nil {
		log.Ctx(ctx).Error().Err(err).Str("template_id", templateID).Msg("failed to fetch additional students")
		return err
	}

	if len(students) == 0 {
		return nil
	}

	queryInsertStudents := `
		INSERT INTO pr_additional_students (
			id,
			pr_id,
			student_id,
			name
		) VALUES (?, ?, ?, ?)
	`

	for _, student := range students {
		_, err := tx.ExecContext(ctx, tx.Rebind(queryInsertStudents),
			ulid.Make().String(), prID, student.StudentID, student.Name,
		)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Str("template_id", templateID).Msg("failed to insert additional students")
			return err
		}
	}

	return nil
}
