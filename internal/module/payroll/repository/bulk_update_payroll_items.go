package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
	"fmt"
	"strings"

	"github.com/oklog/ulid/v2"
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
		if item.IsITP != nil {
			setParts = append(setParts, "is_itp = :is_itp")
			args["is_itp"] = *item.IsITP
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

		if len(setParts) > 0 {
			query := fmt.Sprintf("UPDATE payroll_items SET %s, updated_at = NOW() WHERE id = :id", strings.Join(setParts, ", "))

			_, err := tx.NamedExecContext(ctx, query, args)
			if err != nil {
				return err
			}
		}

		if item.AdditionalStudents != nil {
			queryDelete := `DELETE FROM payroll_item_addtional_students WHERE payroll_item_id = $1`
			if _, err := tx.ExecContext(ctx, queryDelete, item.ID); err != nil {
				return err
			}

			queryInsert := `
				INSERT INTO payroll_item_addtional_students (
					id, payroll_item_id, student_id, name, created_at, updated_at
				) VALUES ($1, $2, $3, $4, NOW(), NOW())
			`
			for _, student := range item.AdditionalStudents {
				_, err := tx.ExecContext(ctx, queryInsert, ulid.Make().String(), item.ID, student.StudentID, student.Name)
				if err != nil {
					return err
				}
			}

			if len(setParts) == 0 {
				if _, err := tx.ExecContext(ctx, "UPDATE payroll_items SET updated_at = NOW() WHERE id = $1", item.ID); err != nil {
					return err
				}
			}
		}
	}

	return tx.Commit()
}
