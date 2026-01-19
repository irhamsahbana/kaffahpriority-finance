package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"

	"github.com/jmoiron/sqlx"
)

func (r *payrollRepo) GetAdditionalStudents(ctx context.Context, payrollItemIDs []string) (map[string][]entity.PayrollItemAdditionalStudent, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetAdditionalStudents")
	defer span.End()

	if len(payrollItemIDs) == 0 {
		return make(map[string][]entity.PayrollItemAdditionalStudent), nil
	}

	query, args, err := sqlx.In(`
		SELECT *
		FROM payroll_item_addtional_students
		WHERE payroll_item_id IN (?)
	`, payrollItemIDs)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)
	var students []entity.PayrollItemAdditionalStudent
	if err := r.db.SelectContext(ctx, &students, query, args...); err != nil {
		return nil, err
	}

	result := make(map[string][]entity.PayrollItemAdditionalStudent)
	for _, s := range students {
		result[s.PayrollItemID] = append(result[s.PayrollItemID], s)
	}

	return result, nil
}
