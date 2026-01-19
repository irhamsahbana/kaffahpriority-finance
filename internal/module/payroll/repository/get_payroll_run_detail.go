package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

func (r *payrollRepo) GetPayrollRunDetail(ctx context.Context, id string) (*entity.GetPayrollRunDetailResp, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetPayrollRunDetail")
	defer span.End()

	// 1. Get Payroll Run
	var run entity.PayrollRun
	queryRun := `
		SELECT *
		FROM payroll_runs
		WHERE id = $1
	`
	if err := r.db.GetContext(ctx, &run, queryRun, id); err != nil {
		return nil, err
	}

	// 2. Get Payroll Items
	var items []entity.PayrollItem
	queryItems := `
		SELECT *
		FROM payroll_items
		WHERE payroll_run_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
	`
	if err := r.db.SelectContext(ctx, &items, queryItems, id); err != nil {
		return nil, err
	}

	// 3. Get Additional Students if items exist
	if len(items) > 0 {
		var itemIDs []string
		for _, item := range items {
			itemIDs = append(itemIDs, item.ID)
		}

		queryAddStudents, args, err := sqlx.In(`
			SELECT *
			FROM payroll_item_addtional_students
			WHERE payroll_item_id IN (?)
		`, itemIDs)
		if err != nil {
			return nil, err
		}

		queryAddStudents = r.db.Rebind(queryAddStudents)
		var additionalStudents []entity.PayrollItemAdditionalStudent
		if err := r.db.SelectContext(ctx, &additionalStudents, queryAddStudents, args...); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to fetch payroll item additional students")
			return nil, err
		}

		// Map additional students to items
		studentsMap := make(map[string][]entity.PayrollItemAdditionalStudent)
		for _, s := range additionalStudents {
			studentsMap[s.PayrollItemID] = append(studentsMap[s.PayrollItemID], s)
		}

		for i := range items {
			if students, ok := studentsMap[items[i].ID]; ok {
				items[i].AdditionalStudents = students
			} else {
				items[i].AdditionalStudents = make([]entity.PayrollItemAdditionalStudent, 0)
			}
		}
	} else {
		items = []entity.PayrollItem{}
	}

	return &entity.GetPayrollRunDetailResp{
		PayrollRun: run,
		Items:      items,
	}, nil
}
