package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/pkg/errmsg"
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (r *reportRepo) UpdateRegistrationV2(ctx context.Context, req *entity.UpdateRegistrationReq) (*entity.UpdateRegistrationResp, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.UpdateRegistrationV2")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("failed to begin transaction")
		return nil, err
	}
	defer func() {
		if err != nil {
			errRB := tx.Rollback()
			if errRB != nil {
				log.Ctx(ctx).Error().Err(errRB).Msgf("failed to rollback transaction")
			}
			return
		}
		errCommit := tx.Commit()
		if errCommit != nil {
			log.Ctx(ctx).Error().Err(errCommit).Msgf("failed to commit transaction")
		}
	}()

	currentReg, err := r.GetRegistration(ctx, &entity.GetRegistrationReq{ID: req.ID})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to get registration")
		return nil, err
	}

	shouldResetFees, err := r.checkResetFees(ctx, currentReg, req)
	if err != nil {
		return nil, err
	}

	if err := r.updateRegistrationRecord(ctx, tx, req, shouldResetFees); err != nil {
		return nil, err
	}

	if err := r.updateRegistrationStudents(ctx, tx, req); err != nil {
		return nil, err
	}

	if err := r.handleTemplateUpdate(ctx, tx, currentReg, req); err != nil {
		return nil, err
	}

	resp := new(entity.UpdateRegistrationResp)
	resp.ID = req.ID

	return resp, nil
}

func (r *reportRepo) checkResetFees(ctx context.Context, currentReg *entity.GetRegistrationResp, req *entity.UpdateRegistrationReq) (bool, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.checkResetFees")
	defer span.End()

	// Check if reset is needed
	lecturerChanged := false
	// lecturer_id is changed from NULL to non-NULL
	if (currentReg.LecturerID == nil && req.LecturerId != nil) ||
		// lecturer_id is changed from non-NULL to NULL
		(currentReg.LecturerID != nil && req.LecturerId == nil) ||
		// lecturer_id is changed from non-NULL to non-NULL with different value
		(currentReg.LecturerID != nil && req.LecturerId != nil && *currentReg.LecturerID != *req.LecturerId) {
		lecturerChanged = true
	}

	// specify timezone
	loc, err := time.LoadLocation("Asia/Makassar")
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("failed to load location")
		return false, err
	}

	var currentRegistrationAllocatedAtTime time.Time
	if currentReg.AllocatedAt != nil && *currentReg.AllocatedAt != "" {
		// convert from string to time
		currentRegistrationAllocatedAtTime, err = time.ParseInLocation(time.RFC3339Nano, *currentReg.AllocatedAt, loc)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to parse allocated_at")
			return false, err
		}
	}

	// Compare months (YYYY-MM)
	currentMonth := currentRegistrationAllocatedAtTime.In(loc).Format("2006-01")
	newMonth := req.AllocatedAt[0:7]
	monthChanged := currentMonth != newMonth

	return lecturerChanged || monthChanged, nil
}

func (r *reportRepo) updateRegistrationRecord(ctx context.Context, tx *sqlx.Tx, req *entity.UpdateRegistrationReq, shouldResetFees bool) error {
	ctx, span := tracing.StartSpan(ctx, "repo.updateRegistrationRecord")
	defer span.End()

	query := `
		UPDATE program_registrations SET
			lecturer_id = ?,
			marketer_id = ?,
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

	// specify timezone
	loc, err := time.LoadLocation("Asia/Makassar")
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("failed to load location")
		return err
	}

	// Combine date and time strings, then parse as timestamp
	paidAtDateTime := req.PaidAt + " " + req.PaidAtTime
	allocatedAtDateTime := req.AllocatedAt + " " + req.AllocatedAtTime

	// parse date
	parsedPaidAt, err := time.ParseInLocation("2006-01-02 15:04:05", paidAtDateTime, loc)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("failed to parse paid_at")
		return err
	}

	parsedAllocatedAt, err := time.ParseInLocation("2006-01-02 15:04:05", allocatedAtDateTime, loc)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("failed to parse allocated_at")
		return err
	}

	_, err = tx.ExecContext(ctx, tx.Rebind(query),
		req.LecturerId, req.MarketerId,
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
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to update data")
		return err
	}
	return nil
}

func (r *reportRepo) updateRegistrationStudents(ctx context.Context, tx *sqlx.Tx, req *entity.UpdateRegistrationReq) error {
	ctx, span := tracing.StartSpan(ctx, "repo.updateRegistrationStudents")
	defer span.End()
	query := `
		DELETE FROM pr_additional_students WHERE pr_id = ?
	`
	_, err := tx.ExecContext(ctx, tx.Rebind(query), req.ID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to delete additional students")
		return err
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
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to insert additional students")
			return err
		}
	}
	return nil
}

func (r *reportRepo) handleTemplateUpdate(ctx context.Context, tx *sqlx.Tx, currentReg *entity.GetRegistrationResp, req *entity.UpdateRegistrationReq) error {
	ctx, span := tracing.StartSpan(ctx, "repo.handleTemplateUpdate")
	defer span.End()
	if !req.IsUpdateTemplate {
		return nil
	}

	// Check if core template fields (lecturer_id, program_id, student_id) have changed
	// If any of these fields changed, create a new template
	// If none changed, update the existing template
	currentTemplate, err := r.GetTemplate(ctx, &entity.GetTemplateReq{
		ID: currentReg.TemplateID,
	})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to get template core fields")
		return err
	}

	// Check if any core field has changed
	lecturerIdChanged := false
	// lecturer_id is changed from NULL to non-NULL
	if (currentTemplate.LecturerId == nil && req.LecturerId != nil) ||
		// lecturer_id is changed from non-NULL to NULL
		(currentTemplate.LecturerId != nil && req.LecturerId == nil) {
		lecturerIdChanged = true
	}

	changeLecturer := currentTemplate.LecturerId != nil && req.LecturerId != nil && *currentTemplate.LecturerId != *req.LecturerId
	if changeLecturer {
		return errmsg.NewCustomErrors(http.StatusUnprocessableEntity).SetMessage("Tidak dapat mengubah mentor atau sudah ada registrasi yang menggunakan mentor tersebut, silahkan buat bank data baru")
	}

	coreFieldsChanged := lecturerIdChanged

	if coreFieldsChanged {
		newTemplateId, err := r.createNewTemplate(ctx, tx, req)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to create new template")
			return err
		}

		// update all registrations that using old template to use new template_id
		err = r.updateRegistrationsTemplateID(ctx, tx, currentReg.TemplateID, newTemplateId)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to update registrations template_id")
			return err
		}

		// Delete old template
		err = r.deleteOldTemplate(ctx, tx, currentReg.TemplateID)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to delete old template")
			return err
		}
	} else {
		err = r.updateExistingTemplate(ctx, tx, currentReg.TemplateID, req)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to update existing template")
			return err
		}
	}

	return nil
}

func (r *reportRepo) updateRegistrationsTemplateID(ctx context.Context, tx *sqlx.Tx, oldTemplateID, newTemplateID string) error {
	ctx, span := tracing.StartSpan(ctx, "repo.updateRegistrationsTemplateID")
	defer span.End()
	query := `
		UPDATE program_registrations SET
			template_id = ?
		WHERE
			template_id = ?
			AND deleted_at IS NULL
	`
	_, err := tx.ExecContext(ctx, tx.Rebind(query), newTemplateID, oldTemplateID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("oldTemplateID", oldTemplateID).Msgf("failed to update registrations template_id")
		return err
	}
	return nil
}

func (r *reportRepo) createNewTemplate(ctx context.Context, tx *sqlx.Tx, req *entity.UpdateRegistrationReq) (string, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.createNewTemplate")
	defer span.End()

	newTemplateId := ulid.Make().String()
	queryProgram := `
		SELECT
			p.price_per_meeting AS program_fee_per_meeting,
			p.commission_fee AS marketer_commission_fee
		FROM
			programs p
		WHERE
			p.id = ?
			AND p.deleted_at IS NULL
	`

	var program struct {
		ProgramFeePerMeeting  float64 `db:"program_fee_per_meeting"`
		MarketerCommissionFee float64 `db:"marketer_commission_fee"`
	}

	err := tx.GetContext(ctx, &program, tx.Rebind(queryProgram), req.ProgramId)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Ctx(ctx).Warn().Err(err).Any("req", req).Msgf("program not found")
			return newTemplateId, errmsg.NewCustomErrors(404).SetMessage("program tidak ditemukan")
		}
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to get program data")
		return newTemplateId, err
	}

	query := `
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
			?,
			?, ?, ?, ?,
			?,
			?, ?, ?, ?, ?
		)
	`

	_, err = tx.ExecContext(ctx, tx.Rebind(query),
		newTemplateId, req.UserID, req.ProgramId, req.LecturerId, req.MarketerId, req.StudentId,
		pq.Array(req.Days), req.Notes, req.ProgramFee,
		program.ProgramFeePerMeeting,
		req.AdministrationFee, req.FLFee, req.NLFee, req.IsITP,
		program.MarketerCommissionFee,
		req.OverpaymentFee, req.HRFee, req.MarketerGiftsFee,
		req.ClosingFeeForOffice, req.ClosingFeeForReward,
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to create new template")
		return newTemplateId, err
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
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to insert additional students into new template")
			return newTemplateId, err
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
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to update registration template_id")
		return newTemplateId, err
	}

	return newTemplateId, nil
}

func (r *reportRepo) updateExistingTemplate(ctx context.Context, tx *sqlx.Tx, templateID string, req *entity.UpdateRegistrationReq) error {
	ctx, span := tracing.StartSpan(ctx, "repo.updateExistingTemplate")
	defer span.End()
	// Update existing template
	// Only update fields that exist in program_registration_templates table
	// Skip fields like notes_for_category which only exist in program_registrations
	query := `
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

	_, err := tx.ExecContext(ctx, tx.Rebind(query),
		req.ProgramId, req.LecturerId, req.MarketerId, req.StudentId,
		req.ProgramId, req.ProgramFee, req.AdministrationFee, req.FLFee, req.NLFee,
		req.MarketerCommissionFee, req.OverpaymentFee, req.HRFee, req.MarketerGiftsFee,
		req.ClosingFeeForOffice, req.ClosingFeeForReward, pq.Array(req.Days), req.Notes,
		req.IsITP,
		templateID,
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to update template data")
		return err
	}

	query = `
		DELETE FROM prt_additional_students WHERE prt_id = ?
	`
	_, err = tx.ExecContext(ctx, tx.Rebind(query), templateID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to delete additional students from template")
		return err
	}

	for _, item := range req.Students {
		query = `
			INSERT INTO prt_additional_students (
				id, prt_id, student_id, name
			) VALUES (?, ?, ?, ?)
		`

		_, err = tx.ExecContext(ctx, tx.Rebind(query),
			ulid.Make().String(), templateID, item.StudentID, item.Name,
		)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to insert additional students into template")
			return err
		}
	}
	return nil
}

func (r *reportRepo) deleteOldTemplate(ctx context.Context, tx *sqlx.Tx, templateID string) error {
	ctx, span := tracing.StartSpan(ctx, "repo.deleteOldTemplate")
	defer span.End()

	// Delete old template
	query := `
		UPDATE program_registration_templates SET
			deleted_at = NOW()
		WHERE
			id = ?
			AND deleted_at IS NULL
	`
	_, err := tx.ExecContext(ctx, tx.Rebind(query), templateID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("templateID", templateID).Msgf("failed to delete old template")
		return err
	}
	return nil
}
