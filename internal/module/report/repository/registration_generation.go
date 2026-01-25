package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (r *reportRepo) GenerateRegistrationReports(ctx context.Context, req *entity.GenerateRegistrationsReq) error {
	ctx, span := tracing.StartSpan(ctx, "repo.GenerateRegistrationReports")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("failed to begin transaction")
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
		log.Ctx(ctx).Error().Err(err).Msgf("failed to select templates")
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
		log.Ctx(ctx).Error().Err(err).Msgf("failed to select deleted templates")
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
			log.Ctx(ctx).Error().Err(err).Msgf("failed to delete registrations")
			return err
		}
	}

	batchId := ulid.Make().String()
	const chunkSize = 500
	totalTemplates := len(templateIds)
	totalChunks := (totalTemplates + chunkSize - 1) / chunkSize

	for start := 0; start < len(templateIds); start += chunkSize {
		end := start + chunkSize
		if end > len(templateIds) {
			end = len(templateIds)
		}

		chunk := templateIds[start:end]
		log.Ctx(ctx).Info().
			Str("batch_id", batchId).
			Int("chunk_total", totalChunks).
			Int("chunk_size", len(chunk)).
			Int("total_templates", totalTemplates).
			Msg("processing registration generation chunk")
		values := make([]string, 0, len(chunk))
		args := make([]any, 0, len(chunk)*2+6)
		for _, templateId := range chunk {
			values = append(values, "(?, ?)")
			args = append(args, templateId, ulid.Make().String())
		}

		queryInsert := `
			WITH mapping(template_id, registration_id) AS (
				VALUES ` + strings.Join(values, ",") + `
			),
			template AS (
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
					p.price_per_meeting AS program_fee_per_meeting,
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
				JOIN
					mapping m
					ON m.template_id = prt.id
				WHERE
					prt.deleted_at IS NULL
			),
			inserted AS (
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
					notes_for_lecturer_wage,
					allocated_at
				)
				SELECT
					m.registration_id,
					t.id,
					?,
					t.program_id,
					t.lecturer_id,
					t.marketer_id,
					t.student_id,
					?,
					t.program_name,
					t.program_fee,
					t.program_fee_per_meeting,
					t.full_fee,
					0,
					(CASE WHEN t.is_itp THEN 2 ELSE 1 END * t.acquisition_rights),
					t.administration_fee,
					t.foreign_learning_fee,
					t.night_learning_fee,
					t.is_itp,
					t.marketer_commission_fee,
					t.overpayment_fee,
					t.hr_fee,
					(t.hr_fee - (40000 * CASE WHEN t.is_itp THEN 2 ELSE 1 END * t.acquisition_rights)),
					NULL,
					(40000 * CASE WHEN t.is_itp THEN 2 ELSE 1 END * t.acquisition_rights),
					t.marketer_gifts_fee,
					t.closing_fee_for_office,
					t.closing_fee_for_reward,
					t.days,
					t.notes,
					(
						SELECT
							COALESCE(prr.notes_for_lecturer_wage, '')
						FROM
							program_registrations prr
						WHERE
							prr.program_id = t.program_id
							AND prr.lecturer_id = t.lecturer_id
							AND prr.student_id = t.student_id
							AND prr.deleted_at IS NULL
						ORDER BY
							prr.id DESC
							LIMIT 1
					),
					NOW()
				FROM
					template t
				JOIN
					mapping m
					ON m.template_id = t.id
				WHERE
					NOT EXISTS (
						SELECT 1
						FROM program_registrations pr
						WHERE
							pr.deleted_at IS NULL
							AND (
								(pr.lecturer_id IS NULL AND t.lecturer_id IS NULL)
								OR (pr.lecturer_id = t.lecturer_id)
							)
							AND pr.student_id = t.student_id
							AND pr.program_id = t.program_id
							AND pr.allocated_at AT TIME ZONE ? >= date_trunc('month', NOW() AT TIME ZONE ?)
							AND pr.allocated_at AT TIME ZONE ? < date_trunc('month', NOW() AT TIME ZONE ?) + interval '1 month'
					)
				RETURNING id, template_id
			)
			SELECT id, template_id FROM inserted
		`

		args = append(args, req.UserID, batchId, req.Timezone, req.Timezone, req.Timezone, req.Timezone)
		rows, err := tx.QueryxContext(ctx, r.db.Rebind(queryInsert), args...)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msg("failed to insert program registrations")
			return err
		}

		type insertedRow struct {
			RegistrationID string `db:"id"`
			TemplateID     string `db:"template_id"`
		}
		inserted := make([]insertedRow, 0)
		for rows.Next() {
			var row insertedRow
			if err := rows.StructScan(&row); err != nil {
				_ = rows.Close()
				log.Ctx(ctx).Error().Err(err).Any("req", req).Msg("failed to scan inserted registrations")
				return err
			}
			inserted = append(inserted, row)
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msg("failed to read inserted registrations")
			return err
		}
		_ = rows.Close()

		if len(inserted) == 0 {
			continue
		}

		templateIDsInserted := make([]string, 0, len(inserted))
		registrationByTemplate := make(map[string]string, len(inserted))
		for _, row := range inserted {
			templateIDsInserted = append(templateIDsInserted, row.TemplateID)
			registrationByTemplate[row.TemplateID] = row.RegistrationID
		}

		queryStudents, argsStudents, err := sqlx.In(`
			SELECT
				prt_id,
				student_id,
				name
			FROM prt_additional_students
			WHERE prt_id IN (?)
		`, templateIDsInserted)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msg("failed to build additional students query")
			return err
		}
		queryStudents = r.db.Rebind(queryStudents)

		type additionalStudent struct {
			TemplateID string `db:"prt_id"`
			StudentID  string `db:"student_id"`
			Name       string `db:"name"`
		}
		var students []additionalStudent
		if err := tx.SelectContext(ctx, &students, queryStudents, argsStudents...); err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msg("failed to select additional students")
			return err
		}

		const studentChunkSize = 500
		for s := 0; s < len(students); s += studentChunkSize {
			e := s + studentChunkSize
			if e > len(students) {
				e = len(students)
			}

			values = values[:0]
			args = args[:0]
			for _, student := range students[s:e] {
				registrationID := registrationByTemplate[student.TemplateID]
				if registrationID == "" {
					continue
				}
				values = append(values, "(?, ?, ?, ?)")
				args = append(args, ulid.Make().String(), registrationID, student.StudentID, student.Name)
			}
			if len(values) == 0 {
				continue
			}

			queryInsertStudents := `
				INSERT INTO pr_additional_students (
					id,
					pr_id,
					student_id,
					name
				) VALUES ` + strings.Join(values, ",")

			if _, err := tx.ExecContext(ctx, r.db.Rebind(queryInsertStudents), args...); err != nil {
				log.Ctx(ctx).Error().Err(err).Any("req", req).Msg("failed to insert additional students")
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("failed to commit transaction")
		return err
	}

	fmt.Printf("[SUCCESS] Finished generating reports.\n")
	return nil
}
