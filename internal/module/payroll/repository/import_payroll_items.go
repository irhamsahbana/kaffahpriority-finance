package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
	"strings"

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

	const chunkSize = 500
	var rowsAffected int64
	for start := 0; start < len(req.Items); start += chunkSize {
		end := start + chunkSize
		if end > len(req.Items) {
			end = len(req.Items)
		}

		items := req.Items[start:end]
		values := make([]string, 0, len(items))
		args := make([]any, 0, len(items)*9)
		for _, item := range items {
			var fl *float64
			if item.ForeignLearningFee != nil {
				val := item.ForeignLearningFee.InexactFloat64()
				fl = &val
			}
			var nl *float64
			if item.NightLearningFee != nil {
				val := item.NightLearningFee.InexactFloat64()
				nl = &val
			}

			values = append(values, "(?::text, ?::integer, ?::numeric, ?::boolean, ?::numeric, ?::numeric, ?::numeric, ?::numeric)")
			args = append(args,
				item.ID,
				item.ProgramMeetings,
				item.InitialWage.InexactFloat64(),
				item.IsMeetingFull,
				fl,
				nl,
				item.Wage.InexactFloat64(),
				item.FullWage.InexactFloat64(),
			)
		}

		query := `
			UPDATE payroll_items AS p
			SET
				updated_at = NOW(),
				program_meetings = v.program_meetings,
				initial_wage = v.initial_wage,
				is_meeting_full = v.is_meeting_full,
				foreign_learning_fee = v.foreign_learning_fee,
				night_learning_fee = v.night_learning_fee,
				wage = v.wage,
				full_wage = v.full_wage
			FROM (
				VALUES ` + strings.Join(values, ",") + `
			) AS v(
				id,
				program_meetings,
				initial_wage,
				is_meeting_full,
				foreign_learning_fee,
				night_learning_fee,
				wage,
				full_wage
			)
			WHERE
				p.id = v.id
				AND p.deleted_at IS NULL
				AND (
					p.program_meetings IS DISTINCT FROM v.program_meetings
					OR p.is_meeting_full IS DISTINCT FROM v.is_meeting_full
					OR p.foreign_learning_fee IS DISTINCT FROM v.foreign_learning_fee
					OR p.night_learning_fee IS DISTINCT FROM v.night_learning_fee
					OR p.wage IS DISTINCT FROM v.wage
					OR p.full_wage IS DISTINCT FROM v.full_wage
				)
		`

		res, execErr := tx.ExecContext(ctx, r.db.Rebind(query), args...)
		if execErr != nil {
			log.Ctx(ctx).Error().Err(execErr).Msg("failed to bulk update payroll items from import")
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
