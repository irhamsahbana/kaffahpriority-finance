package repository

import (
	"codebase-app/internal/infrastructure/tracing"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *payrollRepo) ResetPayrollRunItems(ctx context.Context, runID string) (int64, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.ResetPayrollRunItems")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to begin transaction")
		return 0, err
	}
	defer tx.Rollback()

	// 1. Hard-delete additional students linked to payroll items of this run
	queryDeleteAdditional := `
		DELETE FROM payroll_item_addtional_students
		WHERE payroll_item_id IN (
			SELECT id FROM payroll_items WHERE payroll_run_id = $1
		)
	`
	_, err = tx.ExecContext(ctx, queryDeleteAdditional, runID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to delete payroll item additional students")
		return 0, err
	}

	// 2. Hard-delete all payroll items for this run (both active and soft-deleted)
	queryDeleteItems := `DELETE FROM payroll_items WHERE payroll_run_id = $1`
	result, err := tx.ExecContext(ctx, queryDeleteItems, runID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to delete payroll items")
		return 0, err
	}

	rowsAffected, _ := result.RowsAffected()

	if err := tx.Commit(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to commit transaction")
		return 0, err
	}

	return rowsAffected, nil
}
