package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
)

func (r *payrollRepo) CreatePayrollItems(ctx context.Context, items []entity.PayrollItem) error {
	ctx, span := tracing.StartSpan(ctx, "repo.CreatePayrollItems")
	defer span.End()

	if len(items) == 0 {
		return nil
	}
	query := `
		INSERT INTO payroll_items (
			id, template_id, payroll_run_id, academic_manager_id, academic_manager_name,
			lecturer_id, lecturer_name, student_id, student_name,
			program_id, program_name, marketer_id, marketer_name,
			foreign_learning_fee, night_learning_fee, is_itp,
			program_meetings, is_meeting_full, wage_per_meeting, initial_wage,
			full_wage, wage, acquisition_rights, created_at, updated_at
		) VALUES (
			:id, :template_id, :payroll_run_id, :academic_manager_id, :academic_manager_name,
			:lecturer_id, :lecturer_name, :student_id, :student_name,
			:program_id, :program_name, :marketer_id, :marketer_name,
			:foreign_learning_fee, :night_learning_fee, :is_itp,
			:program_meetings, :is_meeting_full, :wage_per_meeting, :initial_wage,
			:full_wage, :wage, :acquisition_rights, :created_at, :updated_at
		)
	`
	_, err := r.db.NamedExecContext(ctx, query, items)
	return err
}
