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

func (r *reportRepo) CopyRegistrations(ctx context.Context, req *entity.CopyRegistrationsReq) error {
	ctx, span := tracing.StartSpan(ctx, "repo.CopyRegistrations")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to begin transaction")
		return err
	}
	defer func() {
		if err != nil {
			if errRB := tx.Rollback(); errRB != nil {
				log.Ctx(ctx).Error().Err(errRB).Any("req", req).Msg("failed to rollback transaction")
			}
			return
		}
		if errCommit := tx.Commit(); errCommit != nil {
			log.Ctx(ctx).Error().Err(errCommit).Any("req", req).Msg("failed to commit transaction")
		}
	}()

	for _, item := range req.Registrations {
		err = r.processCopyRegistration(ctx, tx, req.UserID, item)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *reportRepo) processCopyRegistration(ctx context.Context, tx *sqlx.Tx, userID string, item entity.CopyRegisItem) error {
	ctx, span := tracing.StartSpan(ctx, "repo.processCopyRegistration")
	defer span.End()

	if err := r.checkCopyRegistrationExists(ctx, tx, item); err != nil {
		return err
	}

	prID := ulid.Make().String()
	if err := r.insertCopyRegistration(ctx, tx, userID, item, prID); err != nil {
		return err
	}

	if err := r.copyAdditionalStudentsForRegistration(ctx, tx, item.RegisId, prID); err != nil {
		return err
	}

	return nil
}

func (r *reportRepo) checkCopyRegistrationExists(ctx context.Context, tx *sqlx.Tx, item entity.CopyRegisItem) error {
	ctx, span := tracing.StartSpan(ctx, "repo.checkCopyRegistrationExists")
	defer span.End()

	// Get template_id and allocated_at month from registration
	var regisInfo struct {
		TemplateID string `db:"template_id"`
	}
	queryGetInfo := `SELECT template_id FROM program_registrations WHERE id = ?`
	if err := tx.GetContext(ctx, &regisInfo, tx.Rebind(queryGetInfo), item.RegisId); err != nil {
		return err
	}

	targetMonth := item.AllocatedAt
	if len(targetMonth) > 7 {
		targetMonth = targetMonth[:7]
	}

	collisionID, paidAt, _, err := r.checkAllocationCollision(ctx, tx, regisInfo.TemplateID, targetMonth, "")
	if err != nil {
		return err
	}

	if collisionID != nil {
		// Format allocation from YYYY-MM to MMMM YYYY
		allocationFormatted := targetMonth
		if len(targetMonth) == 7 { // YYYY-MM format
			// convert month to full month name
			monthNames := map[string]string{
				"01": "Januari", "02": "Februari", "03": "Maret",
				"04": "April", "05": "Mei", "06": "Juni",
				"07": "Juli", "08": "Agustus", "09": "September",
				"10": "Oktober", "11": "November", "12": "Desember",
			}
			month := targetMonth[5:7]
			year := targetMonth[0:4]
			if monthName, ok := monthNames[month]; ok {
				allocationFormatted = monthName + " " + year
			}
		}

		paidAtStr := "tanpa tanggal pembayaran"
		if paidAt.Valid {
			paidAtStr = paidAt.Time.Format("02 January 2006 15:04")
		}

		return errmsg.NewCustomErrors(409).SetMessage("Alokasi untuk bulan " + allocationFormatted + " sudah ada untuk template ini (dibayar pada: " + paidAtStr + ")")
	}

	return nil
}

func (_ *reportRepo) insertCopyRegistration(ctx context.Context, tx *sqlx.Tx, userID string, item entity.CopyRegisItem, prID string) error {
	ctx, span := tracing.StartSpan(ctx, "repo.insertCopyRegistration")
	defer span.End()

	query := `
		INSERT INTO program_registrations (
		id,
		template_id,
		user_id,
		program_id,
		is_itp,
		lecturer_id,
		marketer_id,
		student_id,
		program_name,
		program_fee,
		program_meetings,
		program_fee_per_meeting,
		full_fee,
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
		is_paid,
		allocated_at
		)
		SELECT
			?,
			pr.template_id,
			?,
			pr.program_id,
			pr.is_itp,
			pr.lecturer_id,
			pr.marketer_id,
			pr.student_id,
			pr.program_name,
			pr.program_fee,
			pr.program_meetings,
			pr.program_fee_per_meeting,
			pr.full_fee,
			FLOOR(COALESCE(pr.hr_detail_fee, 0) / 40000),
			pr.foreign_learning_fee,
			pr.night_learning_fee,
			pr.marketer_commission_fee,
			pr.overpayment_fee,
			pr.hr_fee,
			pr.mentor_detail_fee,
			NULL,
			pr.hr_detail_fee,
			pr.days,
			pr.notes,
			pr.is_paid,
			((?::text || ' 00:00:00')::timestamp AT TIME ZONE ?)::timestamptz
		FROM
			program_registrations pr
		WHERE
			pr.id = ?
			AND pr.deleted_at IS NULL
	`

	_, err := tx.ExecContext(ctx, tx.Rebind(query),
		prID, userID,
		item.AllocatedAt, item.Timezone,
		item.RegisId,
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("registration_id", item.RegisId).Msg("failed to insert data")
		return err
	}

	return nil
}

func (_ *reportRepo) copyAdditionalStudentsForRegistration(ctx context.Context, tx *sqlx.Tx, regisID, prID string) error {
	ctx, span := tracing.StartSpan(ctx, "repo.copyAdditionalStudentsForRegistration")
	defer span.End()

	queryStudents := `
		SELECT
			adds.student_id,
			adds.name
		FROM
			pr_additional_students adds
		WHERE
			adds.pr_id = ?
	`

	var students []entity.AddStudent
	if err := tx.SelectContext(ctx, &students, tx.Rebind(queryStudents), regisID); err != nil {
		log.Ctx(ctx).Error().Err(err).Str("registration_id", regisID).Msg("failed to fetch additional students")
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
			log.Ctx(ctx).Error().Err(err).Str("registration_id", regisID).Msg("failed to insert additional students")
			return err
		}
	}

	return nil
}
