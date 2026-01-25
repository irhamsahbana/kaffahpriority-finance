package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
	"fmt"
	"strings"
)

func (r *payrollRepo) BulkUpdatePayrollItems(ctx context.Context, req *entity.BulkUpdatePayrollItemReq) error {
	ctx, span := tracing.StartSpan(ctx, "repo.BulkUpdatePayrollItems")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	for _, item := range req.Data {
		setParts := []string{}
		args := map[string]interface{}{"id": item.ID}

		if item.ProgramMeetings != nil {
			setParts = append(setParts, "program_meetings = :program_meetings")
			args["program_meetings"] = *item.ProgramMeetings
		}
		if item.IsMeetingFull != nil {
			setParts = append(setParts, "is_meeting_full = :is_meeting_full")
			args["is_meeting_full"] = *item.IsMeetingFull
		}
		if item.ForeignLearningFee != nil {
			setParts = append(setParts, "foreign_learning_fee = :foreign_learning_fee")
			args["foreign_learning_fee"] = *item.ForeignLearningFee
		}
		if item.NightLearningFee != nil {
			setParts = append(setParts, "night_learning_fee = :night_learning_fee")
			args["night_learning_fee"] = *item.NightLearningFee
		}
		if item.Wage != nil {
			setParts = append(setParts, "wage = :wage")
			args["wage"] = *item.Wage
		}
		if item.InitialWage != nil {
			setParts = append(setParts, "initial_wage = :initial_wage")
			args["initial_wage"] = *item.InitialWage
		}
		if item.AcquisitionRights != nil {
			setParts = append(setParts, "acquisition_rights = :acquisition_rights")
			args["acquisition_rights"] = *item.AcquisitionRights
		}

		if len(setParts) == 0 {
			continue
		}

		query := fmt.Sprintf("UPDATE payroll_items SET %s, updated_at = NOW() WHERE id = :id", strings.Join(setParts, ", "))

		_, err := tx.NamedExecContext(ctx, query, args)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
