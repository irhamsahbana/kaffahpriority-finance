package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/pkg/errmsg"
	"context"
	"database/sql"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (r *reportRepo) RegistrationMultiAllocation(ctx context.Context, req *entity.RegistrationMuliAllocationReq) error {
	ctx, span := tracing.StartSpan(ctx, "repo.RegistrationMultiAllocation")
	defer span.End()

	fnName := "repo::RegistrationMultiAllocation"

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to begin transaction", fnName)
		return err
	}
	defer tx.Rollback()

	// check if template exists
	query := `SELECT id, program_id, student_id, lecturer_id FROM program_registration_templates WHERE id = $1 AND deleted_at IS NULL`
	var templateId, programId, studentId, lecturerId string
	if err := tx.QueryRowContext(ctx, tx.Rebind(query), req.TemplateId).Scan(&templateId, &programId, &studentId, &lecturerId); err != nil {
		if err == sql.ErrNoRows {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - template not found", fnName)
			return errmsg.NewCustomErrors(404, errmsg.WithMessage("Bank data tidak ditemukan"))
		}
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to get template id", fnName)
		return err
	}

	// check if any existing registrations with the same template_id have allocated_at matching any of the sent allocations
	for _, allocation := range req.Allocations {
		var paidAt sql.NullTime
		checkQuery := `
			SELECT paid_at
			FROM program_registrations
			WHERE template_id = ?
				AND TO_CHAR(allocated_at AT TIME ZONE 'Asia/Makassar', 'YYYY-MM') = ?
				AND deleted_at IS NULL
			LIMIT 1
		`
		err := tx.QueryRowContext(ctx, tx.Rebind(checkQuery), templateId, allocation).Scan(&paidAt)

		if err == nil {
			// Registration exists, format the error message with paid_at date
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

			log.Ctx(ctx).Error().Any("req", req).Str("allocation", allocation).Time("paid_at", paidAt.Time).Msgf("%s - allocation already exists for this template", fnName)
			return errmsg.NewCustomErrors(400, errmsg.WithMessage("Alokasi untuk bulan "+allocationFormatted+" sudah ada untuk template ini (dibayar pada: "+paidAtStr+")"))
		} else if err != sql.ErrNoRows {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Str("allocation", allocation).Msgf("%s - failed to check existing allocations", fnName)
			return err
		}
	}

	for _, allocation := range req.Allocations {
		// check if registration already exists
		type existingData struct {
			ID     string `db:"id"`
			IsPaid bool   `db:"is_paid"`
		}
		var existing existingData
		err := tx.GetContext(ctx, &existing,
			tx.Rebind(queryCheckMultiAllocation),
			allocation,
			programId,
			studentId,
			lecturerId,
		)

		// if registration already exists, skip to the next allocation
		if err == nil {
			// if registration is already paid, skip to the next allocation
			if existing.IsPaid {
				continue
			}

			// if registration is not paid, update the registration is_paid to TRUE and set paid_at to NOW()
			if _, err := tx.ExecContext(ctx, tx.Rebind(`
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
			); err != nil {
				log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to update registration is paid", fnName)
				return err
			}

			continue
		} else if err != sql.ErrNoRows {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to check existing registration", fnName)
			return err
		}

		// Generate a new ID for the registration
		registrationId := ulid.Make().String()

		// Insert the registration using the template
		if _, err := tx.ExecContext(ctx, tx.Rebind(queryInsertRegistrationMulti),
			templateId, // Template ID

			registrationId, // New registration ID
			req.UserID,     // User ID

			req.PaidAt,     // Paid at date
			req.PaidAtTime, // Paid at time
			allocation,     // Allocation date
		); err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to insert registration", fnName)
			return err
		}

		// insert additional student data if exists

		// fetch additional students from prt_additional_students
		var students = make([]entity.AddStudent, 0)
		err = tx.SelectContext(ctx, &students, tx.Rebind(queryStudents), templateId)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to select students", fnName)
			return err
		}

		// insert into pr_additional_students
		for _, student := range students {
			_, err = tx.ExecContext(ctx, tx.Rebind(queryInsertStudents),
				ulid.Make().String(), registrationId, student.StudentID, student.Name,
			)
			if err != nil {
				log.Ctx(ctx).Error().Err(err).Any("req", req).Any("template_id", templateId).Msgf("%s - failed to insert additional students", fnName)
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("%s - failed to commit transaction", fnName)
		return err
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
		TO_CHAR(allocated_at AT TIME ZONE 'Asia/Makassar', 'YYYY-MM') = ?
		AND program_id = ?
		AND student_id = ?
		AND lecturer_id = ?
		AND deleted_at IS NULL
`
