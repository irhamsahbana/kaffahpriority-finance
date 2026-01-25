package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *payrollRepo) ImportPayrollItems(ctx context.Context, req *entity.ImportPayrollItemsReq) (int64, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.ImportPayrollItems")
	defer span.End()

	// 1. Transaction Start
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query := `
		UPDATE payroll_items
		SET
			updated_at = NOW(),
			program_meetings = :program_meetings,
			initial_wage = :initial_wage,
			is_meeting_full = :is_meeting_full,
			foreign_learning_fee = :foreign_learning_fee,
			night_learning_fee = :night_learning_fee,
			acquisition_rights = :acquisition_rights,
			wage = :wage,
			full_wage = :full_wage
		WHERE
			id = :id
			AND deleted_at IS NULL
	`

	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to prepare payroll items update statement")
		return 0, err
	}
	defer stmt.Close()

	var rowsAffected int64
	for _, item := range req.Items {
		res, execErr := stmt.ExecContext(ctx, item)
		if execErr != nil {
			log.Ctx(ctx).Error().Err(execErr).Msg("failed to update payroll item from import")
			return 0, execErr
		}

		affected, rowsErr := res.RowsAffected()
		if rowsErr != nil {
			log.Ctx(ctx).Error().Err(rowsErr).Msg("failed to get rows affected")
			return 0, rowsErr
		}
		rowsAffected += affected
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return rowsAffected, nil
}
