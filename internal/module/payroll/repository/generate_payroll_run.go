package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

func (r *payrollRepo) GeneratePayrollRun(ctx context.Context, period string, timezone string) (*entity.PayrollRun, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GeneratePayrollRun")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to begin transaction")
		return nil, err
	}
	defer tx.Rollback()

	// 1. Check if payroll run exists
	run, err := r.checkExistingPayrollRun(ctx, tx, period)
	// 2.1. Check if error is not related to no rows found
	// If error is not nil and not related to no rows found, return the error
	if err != nil && err != sql.ErrNoRows {
		log.Ctx(ctx).Error().Err(err).Msg("failed to check existing payroll run")
		return nil, err
	}

	// 2.2. Create Payroll Run if it does not exist
	if err != nil && err == sql.ErrNoRows {
		run, err = r.createPayrollRunEntity(ctx, tx, period, timezone)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
				log.Ctx(ctx).Info().Msg("Race condition detected: payroll run created concurrently. Fetching existing run.")
				_ = tx.Rollback()

				var existingRun entity.PayrollRun
				queryCheck := `SELECT * FROM payroll_runs WHERE period = $1 LIMIT 1`
				if err := r.db.GetContext(ctx, &existingRun, queryCheck, period); err != nil {
					log.Ctx(ctx).Error().Err(err).Msg("failed to fetch existing payroll run after race condition")
					return nil, err
				}
				return &existingRun, nil
			}
			return nil, err
		}
	}

	// 3. Sync with templates (Create new items, Delete invalid items)
	// We only sync if the run is in "draft" status.
	if run.Status == entity.PayrollRunStatusDraft {
		templates, err := r.fetchRegistrationTemplates(ctx, tx, period)
		if err != nil {
			return nil, err
		}

		if err := r.syncPayrollItems(ctx, tx, run.ID, templates, period); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to commit transaction")
		return nil, err
	}

	return run, nil
}

func (_ *payrollRepo) checkExistingPayrollRun(ctx context.Context, tx *sqlx.Tx, period string) (*entity.PayrollRun, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.checkExistingPayrollRun")
	defer span.End()

	var existingRun entity.PayrollRun
	queryCheck := `SELECT id, status FROM payroll_runs WHERE period = $1 LIMIT 1`
	err := tx.GetContext(ctx, &existingRun, queryCheck, period)
	if err != nil {
		return nil, err
	}

	return &existingRun, nil
}

func (_ *payrollRepo) createPayrollRunEntity(ctx context.Context, tx *sqlx.Tx, period string, timezone string) (*entity.PayrollRun, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.createPayrollRunEntity")
	defer span.End()

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to load location")
		return nil, err
	}

	periodTime, err := time.Parse("2006-01", period)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to parse period")
		return nil, err
	}

	year, month, _ := periodTime.Date()
	periodStart := time.Date(year, month, 1, 0, 0, 0, 0, loc)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

	runID := ulid.Make().String()
	run := entity.PayrollRun{
		ID:          runID,
		Period:      period,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Timezone:    timezone,
		Status:      "draft",
		CreatedAt:   time.Now(),
	}

	queryRun := `
		INSERT INTO payroll_runs (id, period_start, period_end, period, timezone, status, created_at)
		VALUES (:id, :period_start, :period_end, :period, :timezone, :status, :created_at)
	`
	_, err = tx.NamedExecContext(ctx, queryRun, run)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to insert payroll run")
		return nil, err
	}

	return &run, nil
}

func (_ *payrollRepo) fetchRegistrationTemplates(ctx context.Context, tx *sqlx.Tx, period string) ([]templateData, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.fetchRegistrationTemplates")
	defer span.End()

	// Parse period to get the last day of the month
	periodTime, err := time.Parse("2006-01", period)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to parse period")
		return nil, err
	}

	year, month, _ := periodTime.Date()
	periodEnd := time.Date(year, month+1, 0, 23, 59, 59, 0, time.Local) // Last day of the month

	var templates []templateData
	queryTemplates := `
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
			p.acquisition_rights
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
			prt.deleted_at IS NULL AND prt.created_at <= $1
	`
	if err := tx.SelectContext(ctx, &templates, queryTemplates, periodEnd); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to fetch templates")
		return nil, err
	}
	return templates, nil
}

func (r *payrollRepo) createPayrollItemsFromTemplates(ctx context.Context, tx *sqlx.Tx, runID string, templates []templateData, previousNotes map[string]string) error {
	ctx, span := tracing.StartSpan(ctx, "repo.createPayrollItemsFromTemplates")
	defer span.End()

	var items []entity.PayrollItem
	var additionalStudents []entity.PayrollItemAdditionalStudent
	var templateIDs []string

	for _, t := range templates {
		templateIDs = append(templateIDs, t.TemplateID)
	}

	// Fetch additional students for all templates
	additionalStudentsMap, err := r.fetchAdditionalStudents(ctx, tx, templateIDs)
	if err != nil {
		return err
	}

	for _, t := range templates {
		wagePerMeeting := t.PricePerMeeting
		fullWage := t.FullFee
		acquisitionRights := t.AcquisitionRights

		if t.IsITP {
			wagePerMeeting = wagePerMeeting.Mul(decimal.NewFromInt(2))
			fullWage = fullWage.Mul(decimal.NewFromInt(2))
			acquisitionRights = acquisitionRights * 2
		}

		notes := ""
		if val, ok := previousNotes[t.TemplateID]; ok {
			notes = val
		}

		item := entity.PayrollItem{
			ID:                  ulid.Make().String(),
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
			Notes:               notes,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}

		items = append(items, item)

		// Add additional students if any
		if students, ok := additionalStudentsMap[t.TemplateID]; ok {
			for _, s := range students {
				additionalStudents = append(additionalStudents, entity.PayrollItemAdditionalStudent{
					ID:            ulid.Make().String(),
					PayrollItemID: item.ID,
					StudentID:     s.StudentID,
					Name:          s.Name,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				})
			}
		}
	}

	queryInsertItems := `
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
	_, err = tx.NamedExecContext(ctx, queryInsertItems, items)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to insert payroll items")
		return err
	}

	if len(additionalStudents) > 0 {
		queryInsertAdditionalStudents := `
			INSERT INTO payroll_item_addtional_students (
				id, payroll_item_id, student_id, name, created_at, updated_at
			) VALUES (
				:id, :payroll_item_id, :student_id, :name, :created_at, :updated_at
			)
		`
		_, err = tx.NamedExecContext(ctx, queryInsertAdditionalStudents, additionalStudents)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to insert payroll item additional students")
			return err
		}
	}

	return nil
}

func (r *payrollRepo) syncPayrollItems(ctx context.Context, tx *sqlx.Tx, runID string, templates []templateData, period string) error {
	ctx, span := tracing.StartSpan(ctx, "repo.syncPayrollItems")
	defer span.End()

	// 1. Fetch existing items to compare
	existingItems, err := r.fetchExistingPayrollItemsLight(ctx, tx, runID)
	if err != nil {
		return err
	}

	// 1b. Fetch template IDs of manually deleted items so we don't re-create them
	deletedTemplateIDs, err := r.fetchDeletedPayrollItemTemplateIDs(ctx, tx, runID)
	if err != nil {
		return err
	}

	// 2. Map existing items for O(1) lookup
	// Key: TemplateID
	existingMap := make(map[string]entity.PayrollItem)
	for _, item := range existingItems {
		existingMap[item.TemplateID] = item
	}

	// 3. Identify items to ADD and DELETE
	templateMap := make(map[string]bool)
	var toAdd []templateData
	var toDeleteIDs []string

	// Find items to ADD (in templates but not in DB, and not previously deleted by user)
	for _, t := range templates {
		templateMap[t.TemplateID] = true
		if _, exists := existingMap[t.TemplateID]; !exists {
			// Skip templates whose payroll items were manually deleted by the user
			if deletedTemplateIDs[t.TemplateID] {
				continue
			}
			toAdd = append(toAdd, t)
		}
	}

	// Find items to DELETE (in DB but not in templates)
	for key, item := range existingMap {
		if !templateMap[key] {
			toDeleteIDs = append(toDeleteIDs, item.ID)
		}
	}

	// 4. Execute Operations
	if len(toDeleteIDs) > 0 {
		log.Ctx(ctx).Info().Int("count", len(toDeleteIDs)).Msg("syncPayrollItems: deleting orphaned items")
		if err := r.deletePayrollItems(ctx, tx, toDeleteIDs); err != nil {
			return err
		}
	}

	if len(toAdd) > 0 {
		log.Ctx(ctx).Info().Int("count", len(toAdd)).Msg("syncPayrollItems: adding new items")

		previousNotes := make(map[string]string)
		if periodTime, err := time.Parse("2006-01", period); err == nil {
			prevPeriod := periodTime.AddDate(0, -1, 0).Format("2006-01")

			var templateIDs []string
			for _, t := range toAdd {
				templateIDs = append(templateIDs, t.TemplateID)
			}

			if notes, err := r.fetchPreviousPeriodNotes(ctx, tx, prevPeriod, templateIDs); err == nil {
				previousNotes = notes
			} else {
				log.Ctx(ctx).Warn().Err(err).Msg("failed to fetch previous period notes")
			}
		}

		if err := r.createPayrollItemsFromTemplates(ctx, tx, runID, toAdd, previousNotes); err != nil {
			return err
		}
	}

	return nil
}

func (r *payrollRepo) fetchPreviousPeriodNotes(ctx context.Context, tx *sqlx.Tx, prevPeriod string, templateIDs []string) (map[string]string, error) {
	if len(templateIDs) == 0 {
		return nil, nil
	}

	var results []struct {
		TemplateID string `db:"template_id"`
		Notes      string `db:"notes"`
	}

	query, args, err := sqlx.In(`
		SELECT pi.template_id, pi.notes
		FROM payroll_items pi
		JOIN payroll_runs pr ON pi.payroll_run_id = pr.id
		WHERE pr.period = ? AND pi.notes != '' AND pi.deleted_at IS NULL AND pi.template_id IN (?)
	`, prevPeriod, templateIDs)
	if err != nil {
		return nil, err
	}

	query = tx.Rebind(query)
	if err := tx.SelectContext(ctx, &results, query, args...); err != nil {
		return nil, err
	}

	notesMap := make(map[string]string)
	for _, res := range results {
		notesMap[res.TemplateID] = res.Notes
	}

	return notesMap, nil
}

func (_ *payrollRepo) fetchExistingPayrollItemsLight(ctx context.Context, tx *sqlx.Tx, runID string) ([]entity.PayrollItem, error) {
	var items []entity.PayrollItem
	query := `
		SELECT id, template_id, student_id, program_id, lecturer_id
		FROM payroll_items
		WHERE payroll_run_id = $1 AND deleted_at IS NULL
	`
	if err := tx.SelectContext(ctx, &items, query, runID); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to fetch existing payroll items")
		return nil, err
	}
	return items, nil
}

// fetchDeletedPayrollItemTemplateIDs returns template_ids of payroll items that were
// manually soft-deleted for this run. This prevents syncPayrollItems from re-creating
// items that the user intentionally removed.
func (_ *payrollRepo) fetchDeletedPayrollItemTemplateIDs(ctx context.Context, tx *sqlx.Tx, runID string) (map[string]bool, error) {
	var templateIDs []string
	query := `
		SELECT template_id
		FROM payroll_items
		WHERE payroll_run_id = $1 AND deleted_at IS NOT NULL
	`
	if err := tx.SelectContext(ctx, &templateIDs, query, runID); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to fetch deleted payroll item template IDs")
		return nil, err
	}

	result := make(map[string]bool, len(templateIDs))
	for _, id := range templateIDs {
		result[id] = true
	}
	return result, nil
}

func (_ *payrollRepo) deletePayrollItems(ctx context.Context, tx *sqlx.Tx, ids []string) error {
	ctx, span := tracing.StartSpan(ctx, "repo.deletePayrollItems")
	defer span.End()

	query, args, err := sqlx.In("UPDATE payroll_items SET deleted_at = NOW() WHERE id IN (?)", ids)
	if err != nil {
		return err
	}
	query = tx.Rebind(query)
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to delete payroll items")
		return err
	}
	return nil
}

func (_ *payrollRepo) fetchAdditionalStudents(ctx context.Context, tx *sqlx.Tx, templateIDs []string) (map[string][]entity.PayrollItemAdditionalStudent, error) {
	if len(templateIDs) == 0 {
		return nil, nil
	}

	var students []struct {
		TemplateID string  `db:"prt_id"`
		StudentID  *string `db:"student_id"`
		Name       string  `db:"name"`
	}

	query, args, err := sqlx.In(`
		SELECT prt_id, student_id, name
		FROM prt_additional_students
		WHERE prt_id IN (?)
	`, templateIDs)
	if err != nil {
		return nil, err
	}

	query = tx.Rebind(query)
	if err := tx.SelectContext(ctx, &students, query, args...); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to fetch additional students")
		return nil, err
	}

	result := make(map[string][]entity.PayrollItemAdditionalStudent)
	for _, s := range students {
		result[s.TemplateID] = append(result[s.TemplateID], entity.PayrollItemAdditionalStudent{
			StudentID: s.StudentID,
			Name:      s.Name,
		})
	}

	return result, nil
}

type templateData struct {
	TemplateID          string          `db:"template_id"`
	AcademicManagerID   string          `db:"academic_manager_id"`
	AcademicManagerName string          `db:"academic_manager_name"`
	LecturerID          string          `db:"lecturer_id"`
	LecturerName        string          `db:"lecturer_name"`
	StudentID           string          `db:"student_id"`
	StudentName         string          `db:"student_name"`
	ProgramID           string          `db:"program_id"`
	ProgramName         string          `db:"program_name"`
	MarketerID          string          `db:"marketer_id"`
	MarketerName        string          `db:"marketer_name"`
	ForeignLearningFee  decimal.Decimal `db:"foreign_learning_fee"`
	NightLearningFee    decimal.Decimal `db:"night_learning_fee"`
	IsITP               bool            `db:"is_itp"`
	PricePerMeeting     decimal.Decimal `db:"price_per_meeting"`
	FullFee             decimal.Decimal `db:"full_fee"`
	AcquisitionRights   uint64          `db:"acquisition_rights"`
}
