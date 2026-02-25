package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
	"database/sql"
	"time"

	"codebase-app/pkg/errmsg"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

func (r *payrollRepo) CreatePayrollItem(ctx context.Context, req *entity.CreatePayrollItemReq) (*entity.CreatePayrollItemResp, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.CreatePayrollItem")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. Get or Create Payroll Run for the period
	run, err := r.checkExistingPayrollRun(ctx, tx, req.Period)
	if err != nil && err != sql.ErrNoRows {
		log.Ctx(ctx).Error().Err(err).Msg("failed to check existing payroll run")
		return nil, err
	}

	if err == sql.ErrNoRows {
		// Create run
		timezone := "Asia/Makassar"
		run, err = r.createPayrollRunEntity(ctx, tx, req.Period, timezone)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to create payroll run")
			return nil, err
		}
	}

	if run.Status != entity.PayrollRunStatusDraft {
		log.Ctx(ctx).Warn().Msg("cannot add item to non-draft payroll run")
		return nil, errmsg.NewCustomErrors(400, errmsg.WithMessage("cannot add item to non-draft payroll run"))
	}

	runID := run.ID

	// 2. Fetch template data
	var t templateData
	queryTemplate := `
		SELECT
			prt.id AS template_id,
			COALESCE(am.id, '') AS academic_manager_id,
			COALESCE(am.name, '') AS academic_manager_name,
			COALESCE(l.id, '') AS lecturer_id,
			COALESCE(l.name, '') AS lecturer_name,
			s.id AS student_id,
			s.name AS student_name,
			p.id AS program_id,
			p.name AS program_name,
			m.id AS marketer_id,
			m.name AS marketer_name,
			COALESCE(prt.foreign_learning_fee, 0) AS foreign_learning_fee,
			COALESCE(prt.night_learning_fee, 0) AS night_learning_fee,
			prt.is_itp,
			p.price_per_meeting,
			p.full_fee,
			p.acquisition_rights,
			prt.created_at AS created_at
		FROM
			program_registration_templates prt
		JOIN
			programs p ON prt.program_id = p.id
		JOIN
			lecturers l ON prt.lecturer_id = l.id
		JOIN
			academic_managers am ON l.academic_manager_id = am.id
		JOIN
			students s ON prt.student_id = s.id
		JOIN
			marketers m ON prt.marketer_id = m.id
		WHERE
			prt.id = $1 AND prt.deleted_at IS NULL
	`
	if err := tx.GetContext(ctx, &t, queryTemplate, req.TemplateID); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to fetch template")
		return nil, err
	}

	periodTime, err := time.Parse("2006-01", req.Period)
	if err != nil {
		return nil, errmsg.NewCustomErrors(400, errmsg.WithMessage("invalid period format"))
	}
	pYear, pMonth, _ := periodTime.Date()
	periodEnd := time.Date(pYear, pMonth+1, 0, 23, 59, 59, 0, time.Local)

	if t.CreatedAt.After(periodEnd) {
		indonesianMonths := []string{
			"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember",
		}
		monthName := indonesianMonths[t.CreatedAt.Month()]
		
		errMsg := "Bank data dibuat pada bulan " + monthName + ". Silakan sesuaikan bulan mulai belajar terlebih dahulu."
		log.Ctx(ctx).Warn().Msg(errMsg)
		return nil, errmsg.NewCustomErrors(400, errmsg.WithMessage(errMsg))
	}

	// 3. Check if payroll item already exists and restore if deleted, or create new
	var existingItemID string
	var existingDeletedAt *time.Time
	queryCheckItem := `SELECT id, deleted_at FROM payroll_items WHERE payroll_run_id = $1 AND template_id = $2 LIMIT 1`
	err = tx.QueryRowContext(ctx, queryCheckItem, runID, req.TemplateID).Scan(&existingItemID, &existingDeletedAt)
	if err != nil && err != sql.ErrNoRows {
		log.Ctx(ctx).Error().Err(err).Msg("failed to check existing payroll item")
		return nil, err
	}

	itemID := ulid.Make().String()

	if existingItemID != "" {
		if existingDeletedAt != nil {
			// Restore it
			queryRestore := `UPDATE payroll_items SET deleted_at = NULL, updated_at = NOW() WHERE id = $1`
			if _, err := tx.ExecContext(ctx, queryRestore, existingItemID); err != nil {
				return nil, err
			}
			itemID = existingItemID
		} else {
			// Already exists and active
			return nil, errmsg.NewCustomErrors(400, errmsg.WithMessage("Data santri ini sudah ada pada periode tersebut"))
		}
	} else {
		// Insert new
		wagePerMeeting := t.PricePerMeeting
		fullWage := t.FullFee
		acquisitionRights := t.AcquisitionRights

		if t.IsITP {
			wagePerMeeting = wagePerMeeting.Mul(decimal.NewFromInt(2))
			fullWage = fullWage.Mul(decimal.NewFromInt(2))
			acquisitionRights = acquisitionRights * 2
		}

		item := entity.PayrollItem{
			ID:                  itemID,
			TemplateID:          t.TemplateID,
			PayrollRunID:        runID,
			AcademicManagerID:   t.AcademicManagerID,
			AcademicManagerName: t.AcademicManagerName,
			LecturerID:          t.LecturerID,
			LecturerName:        t.LecturerName,
			StudentID:           t.StudentID,
			StudentName:         t.StudentName,
			ProgramID:           t.ProgramID,
			ProgramName:         t.ProgramName,
			MarketerID:          t.MarketerID,
			MarketerName:        t.MarketerName,
			ForeignLearningFee:  t.ForeignLearningFee,
			NightLearningFee:    t.NightLearningFee,
			IsITP:               t.IsITP,
			ProgramMeetings:     0,
			IsMeetingFull:       false,
			WagePerMeeting:      wagePerMeeting,
			FullWage:            fullWage,
			Wage:                decimal.Zero,
			InitialWage:         decimal.Zero,
			AcquisitionRights:   acquisitionRights,
			Notes:               "",
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}

		queryInsert := `
			INSERT INTO payroll_items (
				id, template_id, payroll_run_id, academic_manager_id, academic_manager_name,
				lecturer_id, lecturer_name, student_id, student_name,
				program_id, program_name, marketer_id, marketer_name,
				foreign_learning_fee, night_learning_fee, is_itp,
				program_meetings, is_meeting_full, wage_per_meeting,
				full_wage, wage, acquisition_rights, notes, created_at, updated_at
			) VALUES (
				:id, :template_id, :payroll_run_id, :academic_manager_id, :academic_manager_name,
				:lecturer_id, :lecturer_name, :student_id, :student_name,
				:program_id, :program_name, :marketer_id, :marketer_name,
				:foreign_learning_fee, :night_learning_fee, :is_itp,
				:program_meetings, :is_meeting_full, :wage_per_meeting,
				:full_wage, :wage, :acquisition_rights, :notes, :created_at, :updated_at
			)
		`
		if _, err := tx.NamedExecContext(ctx, queryInsert, item); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to insert payroll item")
			return nil, err
		}

		// Insert additional students
		additionalMap, err := r.fetchAdditionalStudents(ctx, tx, []string{req.TemplateID})
		if err == nil && len(additionalMap[req.TemplateID]) > 0 {
			var addStudents []entity.PayrollItemAdditionalStudent
			for _, s := range additionalMap[req.TemplateID] {
				addStudents = append(addStudents, entity.PayrollItemAdditionalStudent{
					ID:            ulid.Make().String(),
					PayrollItemID: itemID,
					StudentID:     s.StudentID,
					Name:          s.Name,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				})
			}
			queryInsertAdd := `
				INSERT INTO payroll_item_addtional_students (
					id, payroll_item_id, student_id, name, created_at, updated_at
				) VALUES (
					:id, :payroll_item_id, :student_id, :name, :created_at, :updated_at
				)
			`
			if _, err := tx.NamedExecContext(ctx, queryInsertAdd, addStudents); err != nil {
				log.Ctx(ctx).Error().Err(err).Msg("failed to insert payroll item additional students")
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to commit transaction")
		return nil, err
	}

	return &entity.CreatePayrollItemResp{ID: itemID}, nil
}
