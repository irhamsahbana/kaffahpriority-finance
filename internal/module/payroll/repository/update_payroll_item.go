package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
	"fmt"
	"strings"

	"github.com/oklog/ulid/v2"
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
	if req.InitialWage != nil {
		setParts = append(setParts, "initial_wage = :initial_wage")
		args["initial_wage"] = *req.InitialWage
	}

	if len(setParts) == 0 && req.AdditionalStudents == nil {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if len(setParts) > 0 {
		query := fmt.Sprintf("UPDATE payroll_items SET %s, updated_at = NOW() WHERE id = :id", strings.Join(setParts, ", "))

		_, err := tx.NamedExecContext(ctx, query, args)
		if err != nil {
			return err
		}
	}

	if req.AdditionalStudents != nil {
		queryDelete := `DELETE FROM payroll_item_addtional_students WHERE payroll_item_id = $1`
		if _, err := tx.ExecContext(ctx, queryDelete, req.ID); err != nil {
			return err
		}

		queryInsert := `
			INSERT INTO payroll_item_addtional_students (
				id, payroll_item_id, student_id, name, created_at, updated_at
			) VALUES ($1, $2, $3, $4, NOW(), NOW())
		`
		for _, student := range req.AdditionalStudents {
			_, err := tx.ExecContext(ctx, queryInsert, ulid.Make().String(), req.ID, student.StudentID, student.Name)
			if err != nil {
				return err
			}
		}

		if len(setParts) == 0 {
			if _, err := tx.ExecContext(ctx, "UPDATE payroll_items SET updated_at = NOW() WHERE id = $1", req.ID); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}
