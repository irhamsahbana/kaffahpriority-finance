package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/pkg/errmsg"
	"context"
	"database/sql"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (r *reportRepo) RegistrationMultiAllocation(ctx context.Context, req *entity.RegistrationMuliAllocationReq) error {
	ctx, span := tracing.StartSpan(ctx, "repo.RegistrationMultiAllocation")
	defer span.End()

	template := req.Template

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to begin transaction")
		return err
	}
	defer tx.Rollback()

	// 1. Check validation (existing allocations)
	err = r.validateAllocations(ctx, tx, req, template)
	if err != nil {
		return err
	}

	// 2. Process allocations
	for _, allocation := range req.Allocations {
		err = r.processAllocation(ctx, tx, req, template, allocation)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("failed to commit transaction")
		return err
	}

	return nil
}

func (r *reportRepo) validateAllocations(ctx context.Context, tx *sqlx.Tx, req *entity.RegistrationMuliAllocationReq, template *entity.GetTemplateResp) error {
	ctx, span := tracing.StartSpan(ctx, "repo.validateAllocations")
	defer span.End()

	for _, allocation := range req.Allocations {
		id, paidAt, isPaid, err := r.checkAllocationCollision(ctx, tx, template.ID, allocation, "")
		if err != nil {
			return err
		}

		if id != nil {
			// Registration exists

			// If is_paid is false, overwrite the registration instead of returning error
			if !isPaid {
				err = r.overwriteRegistration(ctx, tx, req, *id)
				if err != nil {
					return err
				}
				continue // Move to the next allocation
			}

			// If is_paid is true, return error as before
			paidAtStr := "tanpa tanggal pembayaran"
			if paidAt.Valid {
				paidAtStr = paidAt.Time.Format("02 January 2006 15:04")
			}

			// Format allocation from YYYY-MM to MMMM YYYY
			allocationFormatted := allocation
			if len(allocation) == 7 { // YYYY-MM format
				// convert month to full month name
				monthNames := map[string]string{
					"01": "Januari", "02": "Februari", "03": "Maret",
					"04": "April", "05": "Mei", "06": "Juni",
					"07": "Juli", "08": "Agustus", "09": "September",
					"10": "Oktober", "11": "November", "12": "Desember",
				}
				month := allocation[5:7]
				year := allocation[0:4]
				if monthName, ok := monthNames[month]; ok {
					allocationFormatted = monthName + " " + year
				}
			}

			log.Ctx(ctx).Error().Any("req", req).Str("allocation", allocation).Time("paid_at", paidAt.Time).Msgf("allocation already exists for this template")
			return errmsg.NewCustomErrors(http.StatusUnprocessableEntity,
				errmsg.WithMessage("Alokasi untuk bulan "+allocationFormatted+" sudah ada untuk template ini (dibayar pada: "+paidAtStr+")"))
		}
	}
	return nil
}

func (r *reportRepo) checkAllocationCollision(ctx context.Context, tx *sqlx.Tx, templateID string, allocationMonth string, excludeDetailsID string) (*string, sql.NullTime, bool, error) {
	var id string
	var paidAt sql.NullTime
	var isPaid bool

	checkQuery := `
		SELECT id, paid_at, is_paid
		FROM program_registrations
		WHERE template_id = ?
			AND allocated_at AT TIME ZONE 'Asia/Makassar' >= (TO_TIMESTAMP(?, 'YYYY-MM-DD') AT TIME ZONE 'UTC')
			AND allocated_at AT TIME ZONE 'Asia/Makassar' < (TO_TIMESTAMP(?, 'YYYY-MM-DD') AT TIME ZONE 'UTC' + INTERVAL '1 month')
			AND deleted_at IS NULL
	`

	args := []interface{}{templateID, allocationMonth + "-01", allocationMonth + "-01"}

	if excludeDetailsID != "" {
		checkQuery += " AND id != ? "
		args = append(args, excludeDetailsID)
	}

	checkQuery += " LIMIT 1"

	err := tx.QueryRowContext(ctx, tx.Rebind(checkQuery), args...).Scan(&id, &paidAt, &isPaid)

	if err == nil {
		return &id, paidAt, isPaid, nil
	} else if err != sql.ErrNoRows {
		log.Ctx(ctx).Error().Err(err).Str("allocation", allocationMonth).Msgf("failed to check existing allocations")
		return nil, sql.NullTime{}, false, err
	}

	return nil, sql.NullTime{}, false, nil
}

func (r *reportRepo) overwriteRegistration(ctx context.Context, tx *sqlx.Tx, req *entity.RegistrationMuliAllocationReq, registrationId string) error {

	ctx, span := tracing.StartSpan(ctx, "repo.overwriteRegistration")
	defer span.End()

	query := `
		UPDATE program_registrations
		SET
			is_paid = TRUE,
			paid_at = (? || ' ' || ?)::timestamp AT TIME ZONE 'Asia/Makassar',
			updated_at = NOW()
		WHERE id = ?
		`

	_, err := tx.ExecContext(ctx, tx.Rebind(query),
		req.PaidAt,     // Paid at date
		req.PaidAtTime, // Paid at time
		registrationId, // Registration ID
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to update registration is paid")
		return err
	}

	return nil
}

func (r *reportRepo) processAllocation(ctx context.Context, tx *sqlx.Tx, req *entity.RegistrationMuliAllocationReq, template *entity.GetTemplateResp, allocation string) error {
	ctx, span := tracing.StartSpan(ctx, "repo.processAllocation")
	defer span.End()

	// check if registration already exists
	type existingData struct {
		ID     string `db:"id"`
		IsPaid bool   `db:"is_paid"`
	}
	var existing existingData
	allocationDate := allocation + "-01"
	err := tx.GetContext(ctx, &existing,
		tx.Rebind(queryCheckMultiAllocation),
		allocationDate,
		allocationDate,
		template.ProgramId,
		template.StudentId,
		template.LecturerId,
	)

	// if registration already exists, skip to the next allocation
	if err == nil {
		// if registration is already paid, skip to the next allocation
		if existing.IsPaid {
			return nil
		}

		// if registration is not paid, update the registration is_paid to TRUE and set paid_at to NOW()
		_, err = tx.ExecContext(ctx, tx.Rebind(`
			UPDATE program_registrations
			SET
				is_paid = TRUE,
				paid_at = (? || ' ' || ?)::timestamp AT TIME ZONE 'Asia/Makassar',
				updated_at = NOW()
			WHERE id = ?
			`),
			req.PaidAt,     // Paid at date
			req.PaidAtTime, // Paid at time
			existing.ID,    // Registration ID
		)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to update registration is paid")
			return err
		}

		return nil
	} else if err != sql.ErrNoRows {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to check existing registration")
		return err
	}

	// Generate a new ID for the registration
	registrationId := ulid.Make().String()

	// Insert the registration using the template
	_, err = tx.ExecContext(ctx, tx.Rebind(queryInsertRegistrationMulti),
		template.ID, // Template ID

		registrationId, // New registration ID
		req.UserID,     // User ID

		req.PaidAt,     // Paid at date
		req.PaidAtTime, // Paid at time
		allocation,     // Allocation date
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to insert registration")
		return err
	}

	// insert additional student data if exists

	// fetch additional students from prt_additional_students
	var students = make([]entity.AddStudent, 0)
	err = tx.SelectContext(ctx, &students, tx.Rebind(queryStudents), template.ID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to select students")
		return err
	}

	// insert into pr_additional_students
	for _, student := range students {
		_, err = tx.ExecContext(ctx, tx.Rebind(queryInsertStudents),
			ulid.Make().String(), registrationId, student.StudentID, student.Name,
		)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Any("template_id", template.ID).Msgf("failed to insert additional students")
			return err
		}
	}

	return nil
}

var queryInsertRegistrationMulti = `
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
	is_paid,
	paid_at,
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
	(
		(SELECT hr_fee FROM template)
		- (40000 * CASE WHEN (SELECT is_itp FROM template) THEN 2 ELSE 1 END * (SELECT acquisition_rights FROM template))
	),
	NULL,
	(40000 * CASE WHEN (SELECT is_itp FROM template) THEN 2 ELSE 1 END * (SELECT acquisition_rights FROM template)),
	(SELECT marketer_gifts_fee FROM template),
	(SELECT closing_fee_for_office FROM template),
	(SELECT closing_fee_for_reward FROM template),
	(SELECT days FROM template),
	(SELECT notes FROM template),
	TRUE,
	(? || ' ' || ?)::timestamp AT TIME ZONE 'Asia/Makassar',
	(? || '-10 00:00:00')::timestamp AT TIME ZONE 'Asia/Makassar'
`

var queryCheckMultiAllocation = `
	SELECT
		id
	FROM
		program_registrations
	WHERE
		allocated_at AT TIME ZONE 'Asia/Makassar' >= (TO_TIMESTAMP(?, 'YYYY-MM-DD') AT TIME ZONE 'UTC')
		AND allocated_at AT TIME ZONE 'Asia/Makassar' < (TO_TIMESTAMP(?, 'YYYY-MM-DD') AT TIME ZONE 'UTC' + INTERVAL '1 month')
		AND program_id = ?
		AND student_id = ?
		AND lecturer_id = ?
		AND deleted_at IS NULL
`
