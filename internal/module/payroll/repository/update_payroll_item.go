package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
	"fmt"
	"strings"
)

func (r *payrollRepo) UpdatePayrollItem(ctx context.Context, req *entity.UpdatePayrollItemReq) error {
	ctx, span := tracing.StartSpan(ctx, "repo.UpdatePayrollItem")
	defer span.End()

	setParts := []string{}
	args := map[string]interface{}{"id": req.ID}

	if req.ProgramMeetings != nil {
		setParts = append(setParts, "program_meetings = :program_meetings")
		args["program_meetings"] = *req.ProgramMeetings
	}
	if req.IsMeetingFull != nil {
		setParts = append(setParts, "is_meeting_full = :is_meeting_full")
		args["is_meeting_full"] = *req.IsMeetingFull
	}
	if req.ForeignLearningFee != nil {
		setParts = append(setParts, "foreign_learning_fee = :foreign_learning_fee")
		args["foreign_learning_fee"] = *req.ForeignLearningFee
	}
	if req.NightLearningFee != nil {
		setParts = append(setParts, "night_learning_fee = :night_learning_fee")
		args["night_learning_fee"] = *req.NightLearningFee
	}
	if req.Wage != nil {
		setParts = append(setParts, "wage = :wage")
		args["wage"] = *req.Wage
	}

	if len(setParts) == 0 {
		return nil
	}

	query := fmt.Sprintf("UPDATE payroll_items SET %s, updated_at = NOW() WHERE id = :id", strings.Join(setParts, ", "))

	_, err := r.db.NamedExecContext(ctx, query, args)
	return err
}
