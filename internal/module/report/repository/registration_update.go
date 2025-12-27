package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/pkg/errmsg"
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

// getStringValue returns the string value or "<nil>" if pointer is nil
func getStringValue(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func (r *reportRepo) UpdateRegistration(ctx context.Context, req *entity.UpdateRegistrationReq) (*entity.UpdateRegistrationResp, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.UpdateRegistration")
	defer span.End()

	fnName := "repo::UpdateRegistration"
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("%s - failed to begin transaction", fnName)
		return nil, err
	}
	defer func() {
		if err != nil {
			errRB := tx.Rollback()
			if errRB != nil {
				log.Ctx(ctx).Error().Err(errRB).Msgf("%s - failed to rollback transaction", fnName)
			}
			return
		}
		errCommit := tx.Commit()
		if errCommit != nil {
			log.Ctx(ctx).Error().Err(errCommit).Msgf("%s - failed to commit transaction", fnName)
		}
	}()

	// Get current registration data to check if lecturer_id, program_id, or student_id is being changed
	type currentRegistration struct {
		LecturerId  *string   `db:"lecturer_id"`
		ProgramId   string    `db:"program_id"`
		StudentId   string    `db:"student_id"`
		AllocatedAt time.Time `db:"allocated_at"`
	}
	var currentReg currentRegistration
	queryCurrent := `
		SELECT
			lecturer_id,
			program_id,
			student_id,
			allocated_at
		FROM
			program_registrations
		WHERE
			id = ?
			AND deleted_at IS NULL
	`
	err = tx.GetContext(ctx, &currentReg, tx.Rebind(queryCurrent), req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Ctx(ctx).Warn().Err(err).Any("req", req).Msgf("%s - registration not found", fnName)
			return nil, errmsg.NewCustomErrors(404).SetMessage("Registrasi tidak ditemukan")
		}
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to get current registration data", fnName)
		return nil, err
	}

	// // Check if lecturer_id is being changed
	// lecturerIdChanged := false
	// if (currentReg.LecturerId == nil && req.LecturerId != nil) ||
	// 	(currentReg.LecturerId != nil && req.LecturerId == nil) ||
	// 	(currentReg.LecturerId != nil && req.LecturerId != nil && *currentReg.LecturerId != *req.LecturerId) {
	// 	lecturerIdChanged = true
	// }

	// // Check if program_id is being changed
	// programIdChanged := currentReg.ProgramId != req.ProgramId

	// // Check if student_id is being changed
	// studentIdChanged := currentReg.StudentId != req.StudentId

	// // Only check for duplicate if:
	// // 1. Either lecturer_id, program_id, or student_id is being changed
	// // 2. AND the new lecturer_id is not NULL (NULL lecturer_id can have multiple entries per program)
	// if (lecturerIdChanged || programIdChanged || studentIdChanged) && req.LecturerId != nil {
	// 	var count int
	// 	// Use IS NOT DISTINCT FROM to properly handle NULL comparisons
	// 	queryCheck := `
	// 		SELECT
	// 			COUNT(*)
	// 		FROM
	// 			program_registrations
	// 		WHERE
	// 			lecturer_id IS NOT DISTINCT FROM ?
	// 			AND program_id = ?
	// 			AND student_id = ?
	// 			AND id != ?
	// 			AND deleted_at IS NULL
	// 	`
	// 	err = tx.GetContext(ctx, &count, tx.Rebind(queryCheck), req.LecturerId, req.ProgramId, req.StudentId, req.ID)
	// 	if err != nil {
	// 		log.Ctx(ctx).Error().Err(err).
	// 			Str("current_lecturer_id", getStringValue(currentReg.LecturerId)).
	// 			Str("req_lecturer_id", getStringValue(req.LecturerId)).
	// 			Str("current_program_id", currentReg.ProgramId).
	// 			Str("req_program_id", req.ProgramId).
	// 			Str("current_student_id", currentReg.StudentId).
	// 			Str("req_student_id", req.StudentId).
	// 			Bool("lecturer_id_changed", lecturerIdChanged).
	// 			Bool("program_id_changed", programIdChanged).
	// 			Bool("student_id_changed", studentIdChanged).
	// 			Any("req", req).
	// 			Msgf("%s - failed to check duplicate lecturer_id, program_id, and student_id", fnName)
	// 		return nil, err
	// 	}
	// 	if count > 0 {
	// 		log.Ctx(ctx).Warn().
	// 			Str("current_lecturer_id", getStringValue(currentReg.LecturerId)).
	// 			Str("req_lecturer_id", getStringValue(req.LecturerId)).
	// 			Str("current_program_id", currentReg.ProgramId).
	// 			Str("req_program_id", req.ProgramId).
	// 			Str("current_student_id", currentReg.StudentId).
	// 			Str("req_student_id", req.StudentId).
	// 			Int("duplicate_count", count).
	// 			Bool("lecturer_id_changed", lecturerIdChanged).
	// 			Bool("program_id_changed", programIdChanged).
	// 			Bool("student_id_changed", studentIdChanged).
	// 			Any("req", req).
	// 			Msgf("%s - duplicate combination of lecturer_id, program_id, and student_id already exists", fnName)
	// 		return nil, errmsg.NewCustomErrors(409).SetMessage("Kombinasi lecturer_id, program_id, dan student_id sudah ada dalam program_registrations")
	// 	}
	// }

	// specify timezone
	loc, err := time.LoadLocation("Asia/Makassar")
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("%s - failed to load location", fnName)
		return nil, err
	}

	// Check if reset is needed
	lecturerChanged := false
	if (currentReg.LecturerId == nil && req.LecturerId != nil) ||
		(currentReg.LecturerId != nil && req.LecturerId == nil) ||
		(currentReg.LecturerId != nil && req.LecturerId != nil && *currentReg.LecturerId != *req.LecturerId) {
		lecturerChanged = true
	}

	// Compare months (YYYY-MM)
	currentMonth := currentReg.AllocatedAt.In(loc).Format("2006-01")
	newMonth := req.AllocatedAt[0:7]
	monthChanged := currentMonth != newMonth

	shouldResetFees := lecturerChanged || monthChanged

	query := `
		UPDATE program_registrations SET
			program_id = ?,
			lecturer_id = ?,
			marketer_id = ?,
			student_id = ?,
			program_name = (SELECT name FROM programs WHERE id = ?),
			program_fee_per_meeting = (SELECT price_per_meeting FROM programs WHERE id = ?),
			full_fee = (SELECT full_fee FROM programs WHERE id = ?),
			program_acquisition_rights = (CASE WHEN ? THEN 2 ELSE 1 END * (SELECT acquisition_rights FROM programs WHERE id = ?)),
			program_fee = ?,
			administration_fee = ?,
			foreign_learning_fee = ?,
			night_learning_fee = ?,
			marketer_commission_fee = ?,
			overpayment_fee = ?,
			hr_fee = ?,
            mentor_detail_fee = (
                ? - (
                    40000 *
                    CASE
                        WHEN ? THEN 2
                        ELSE 1
                    END *
                    (SELECT acquisition_rights FROM programs WHERE id = ?)
                )
            ),
			hr_detail_fee = (
				40000 *
				CASE
					WHEN ? THEN 2
					ELSE 1
				END *
				(SELECT acquisition_rights FROM programs WHERE id = ?)
			),
			marketer_gifts_fee = ?,
			closing_fee_for_office = ?,
			closing_fee_for_reward = ?,
			days = ?,
			notes = ?,
			notes_for_category = ?,
			is_itp = ?,
			paid_at = ?,
			allocated_at = ?,
			mentor_detail_fee_used = CASE WHEN ? THEN NULL ELSE mentor_detail_fee_used END,
			notes_for_fund_distributions = CASE WHEN ? THEN NULL ELSE notes_for_fund_distributions END,
			updated_at = NOW()
		WHERE
			(id = ? OR parent_id = ?)
			AND deleted_at IS NULL
	`

	// Combine date and time strings, then parse as timestamp
	paidAtDateTime := req.PaidAt + " " + req.PaidAtTime
	allocatedAtDateTime := req.AllocatedAt + " " + req.AllocatedAtTime

	// parse date
	parsedPaidAt, err := time.ParseInLocation("2006-01-02 15:04:05", paidAtDateTime, loc)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("%s - failed to parse paid_at", fnName)
		return nil, err
	}

	parsedAllocatedAt, err := time.ParseInLocation("2006-01-02 15:04:05", allocatedAtDateTime, loc)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("%s - failed to parse allocated_at", fnName)
		return nil, err
	}

	_, err = tx.ExecContext(ctx, tx.Rebind(query),
		req.ProgramId, req.LecturerId, req.MarketerId, req.StudentId,
		req.ProgramId, req.ProgramId, req.ProgramId, req.IsITP, req.ProgramId, req.ProgramFee, req.AdministrationFee, req.FLFee, req.NLFee,
		req.MarketerCommissionFee, req.OverpaymentFee,
		req.HRFee,
		req.HRFee, req.IsITP, req.ProgramId, // calculate mentor_detail_fee
		req.IsITP, req.ProgramId, // calculate hr_detail_fee
		req.MarketerGiftsFee,
		req.ClosingFeeForOffice, req.ClosingFeeForReward, pq.Array(req.Days), req.Notes, req.NotesForCategory,
		req.IsITP,
		parsedPaidAt.Format(time.RFC3339), parsedAllocatedAt.Format(time.RFC3339),
		shouldResetFees, // reset mentor_detail_fee_used
		shouldResetFees, // reset notes_for_fund_distributions
		req.ID,          // for id = ?
		req.ID,          // for parent_id = ?
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to update data", fnName)
		return nil, err
	}

	query = `
		DELETE FROM pr_additional_students WHERE pr_id = ?
	`
	_, err = tx.ExecContext(ctx, tx.Rebind(query), req.ID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to delete additional students", fnName)
		return nil, err
	}

	for _, item := range req.Students {
		query = `
			INSERT INTO pr_additional_students (
				id, pr_id, student_id, name
			) VALUES (?, ?, ?, ?)
		`

		_, err = tx.ExecContext(ctx, tx.Rebind(query),
			ulid.Make().String(), req.ID, item.StudentID, item.Name,
		)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to insert additional students", fnName)
			return nil, err
		}
	}

	if !req.IsUpdateTemplate {
		resp := new(entity.UpdateRegistrationResp)
		resp.ID = req.ID
		return resp, nil
	}

	// update template
	type registration struct {
		TemplateId string `db:"template_id"`
	}
	var reg registration
	query = `
		SELECT
			template_id
		FROM
			program_registrations
		WHERE
			id = ?
			AND deleted_at IS NULL
	`

	err = tx.GetContext(ctx, &reg, tx.Rebind(query), req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Ctx(ctx).Warn().Err(err).Any("req", req).Msgf("%s - registration not found", fnName)
			return nil, errmsg.NewCustomErrors(404).SetMessage("template tidak ditemukan")
		}
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to get registration data", fnName)
		return nil, err
	}

	// Check if core template fields (lecturer_id, program_id, student_id) have changed
	// If any of these fields changed, create a new template
	// If none changed, update the existing template
	type templateCoreFields struct {
		LecturerId *string `db:"lecturer_id"`
		ProgramId  string  `db:"program_id"`
		StudentId  string  `db:"student_id"`
	}
	var currentTemplate templateCoreFields
	queryGetTemplate := `
		SELECT
			lecturer_id,
			program_id,
			student_id
		FROM
			program_registration_templates
		WHERE
			id = ?
			AND deleted_at IS NULL
	`
	err = tx.GetContext(ctx, &currentTemplate, tx.Rebind(queryGetTemplate), reg.TemplateId)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to get template core fields", fnName)
		return nil, err
	}

	// Check if any core field has changed
	lecturerIdChanged := false
	if (currentTemplate.LecturerId == nil && req.LecturerId != nil) ||
		(currentTemplate.LecturerId != nil && req.LecturerId == nil) ||
		(currentTemplate.LecturerId != nil && req.LecturerId != nil && *currentTemplate.LecturerId != *req.LecturerId) {
		lecturerIdChanged = true
	}
	programIdChanged := currentTemplate.ProgramId != req.ProgramId
	studentIdChanged := currentTemplate.StudentId != req.StudentId

	coreFieldsChanged := lecturerIdChanged || programIdChanged || studentIdChanged

	if coreFieldsChanged {
		// Create new template instead of updating existing one
		newTemplateId := ulid.Make().String()

		query = `
			WITH program AS (
				SELECT
					p.price_per_meeting AS program_fee_per_meeting,
					p.commission_fee AS marketer_commission_fee
				FROM
					programs p
				WHERE
					p.id = ?
					AND p.deleted_at IS NULL
			)
			INSERT INTO program_registration_templates (
				id,
				user_id,
				program_id,
				lecturer_id,
				marketer_id,
				student_id,
				days,
				notes,
				program_fee,
				program_fee_per_meeting,
				administration_fee,
				foreign_learning_fee,
				night_learning_fee,
				is_itp,
				marketer_commission_fee,
				overpayment_fee,
				hr_fee,
				marketer_gifts_fee,
				closing_fee_for_office,
				closing_fee_for_reward
			) VALUES (
				?, ?, ?, ?, ?, ?, ?, ?, ?,
				(SELECT program_fee_per_meeting FROM program),
				?, ?, ?, ?,
				(SELECT marketer_commission_fee FROM program),
				?, ?, ?, ?, ?
			)
		`

		_, err = tx.ExecContext(ctx, tx.Rebind(query),
			req.ProgramId,
			newTemplateId, req.UserID, req.ProgramId, req.LecturerId, req.MarketerId, req.StudentId,
			pq.Array(req.Days), req.Notes, req.ProgramFee,
			req.AdministrationFee, req.FLFee, req.NLFee, req.IsITP,
			req.OverpaymentFee, req.HRFee, req.MarketerGiftsFee,
			req.ClosingFeeForOffice, req.ClosingFeeForReward,
		)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to create new template", fnName)
			return nil, err
		}

		// Insert additional students for new template
		for _, item := range req.Students {
			query = `
				INSERT INTO prt_additional_students (
					id, prt_id, student_id, name
				) VALUES (?, ?, ?, ?)
			`

			_, err = tx.ExecContext(ctx, tx.Rebind(query),
				ulid.Make().String(), newTemplateId, item.StudentID, item.Name,
			)
			if err != nil {
				log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to insert additional students into new template", fnName)
				return nil, err
			}
		}

		// Update registration to use new template_id
		query = `
			UPDATE program_registrations SET
				template_id = ?
			WHERE
				id = ?
				AND deleted_at IS NULL
		`
		_, err = tx.ExecContext(ctx, tx.Rebind(query), newTemplateId, req.ID)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to update registration template_id", fnName)
			return nil, err
		}
	} else {
		// Update existing template
		// Only update fields that exist in program_registration_templates table
		// Skip fields like notes_for_category which only exist in program_registrations
		query = `
			UPDATE program_registration_templates SET
				program_id = ?,
				lecturer_id = ?,
				marketer_id = ?,
				student_id = ?,
				program_fee_per_meeting = (SELECT price_per_meeting FROM programs WHERE id = ?),
				program_fee = ?,
				administration_fee = ?,
				foreign_learning_fee = ?,
				night_learning_fee = ?,
				marketer_commission_fee = ?,
				overpayment_fee = ?,
				hr_fee = ?,
				marketer_gifts_fee = ?,
				closing_fee_for_office = ?,
				closing_fee_for_reward = ?,
				days = ?,
				notes = ?,
				is_itp = ?,
				updated_at = NOW()
			WHERE
				id = ?
				AND deleted_at IS NULL
		`

		_, err = tx.ExecContext(ctx, tx.Rebind(query),
			req.ProgramId, req.LecturerId, req.MarketerId, req.StudentId,
			req.ProgramId, req.ProgramFee, req.AdministrationFee, req.FLFee, req.NLFee,
			req.MarketerCommissionFee, req.OverpaymentFee, req.HRFee, req.MarketerGiftsFee,
			req.ClosingFeeForOffice, req.ClosingFeeForReward, pq.Array(req.Days), req.Notes,
			req.IsITP,
			reg.TemplateId,
		)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to update template data", fnName)
			return nil, err
		}

		query = `
			DELETE FROM prt_additional_students WHERE prt_id = ?
		`
		_, err = tx.ExecContext(ctx, tx.Rebind(query), reg.TemplateId)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to delete additional students from template", fnName)
			return nil, err
		}

		for _, item := range req.Students {
			query = `
				INSERT INTO prt_additional_students (
					id, prt_id, student_id, name
				) VALUES (?, ?, ?, ?)
			`

			_, err = tx.ExecContext(ctx, tx.Rebind(query),
				ulid.Make().String(), reg.TemplateId, item.StudentID, item.Name,
			)
			if err != nil {
				log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to insert additional students into template", fnName)
				return nil, err
			}
		}
	}

	resp := new(entity.UpdateRegistrationResp)
	resp.ID = req.ID

	return resp, nil
}
