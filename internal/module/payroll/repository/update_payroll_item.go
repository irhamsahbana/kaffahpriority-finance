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

	if req.LecturerID != nil {
		setParts = append(setParts, "lecturer_id = :lecturer_id")
		args["lecturer_id"] = *req.LecturerID
	}
	if req.LecturerName != nil {
		setParts = append(setParts, "lecturer_name = :lecturer_name")
		args["lecturer_name"] = *req.LecturerName
	}
	if req.ProgramID != nil {
		setParts = append(setParts, "program_id = :program_id")
		args["program_id"] = *req.ProgramID
	}
	if req.ProgramName != nil {
		setParts = append(setParts, "program_name = :program_name")
		args["program_name"] = *req.ProgramName
	}
	if req.ProgramMeetings != nil {
		setParts = append(setParts, "program_meetings = :program_meetings")
		args["program_meetings"] = *req.ProgramMeetings
	}
	if req.AcquisitionRights != nil {
		setParts = append(setParts, "acquisition_rights = :acquisition_rights")
		args["acquisition_rights"] = *req.AcquisitionRights
	}
	if req.Notes != nil {
		setParts = append(setParts, "notes = :notes")
		args["notes"] = *req.Notes
	}
	if req.IsMeetingFull != nil {
		setParts = append(setParts, "is_meeting_full = :is_meeting_full")
		args["is_meeting_full"] = *req.IsMeetingFull
	}
	if req.IsITP != nil {
		setParts = append(setParts, "is_itp = :is_itp")
		args["is_itp"] = *req.IsITP
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
	if req.WagePerMeeting != nil {
		setParts = append(setParts, "wage_per_meeting = :wage_per_meeting")
		args["wage_per_meeting"] = *req.WagePerMeeting
	}
	if req.FullWage != nil {
		setParts = append(setParts, "full_wage = :full_wage")
		args["full_wage"] = *req.FullWage
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
