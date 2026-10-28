package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *reportRepo) ImportLecturersWages(ctx context.Context, req *entity.ImportLecturersWagesReq) error {
	ctx, span := tracing.StartSpan(ctx, "repo.ImportLecturersWages")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to begin transaction")
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE program_registrations
		SET
			updated_at = NOW(),
			program_meetings = ?,
			initial_fee = ?,
			foreign_learning_fee = ?,
			night_learning_fee = ?,
			pc_fee = COALESCE(?, pc_fee),
			mt_fee = COALESCE(?, mt_fee),
			cl_fee = COALESCE(?, cl_fee),
			ms_fee = COALESCE(?, ms_fee),
			sc_fee = COALESCE(?, sc_fee),
			ln_fee = COALESCE(?, ln_fee),
			is_full_fee = ?,
			notes_for_lecturer_wage = ?
		WHERE
			id = ?
			AND deleted_at IS NULL
	`
	query = r.db.Rebind(query)

	for _, data := range req.Registrations {
		if _, err := tx.ExecContext(ctx, query,
			data.ProgramMeetings,
			data.InitialFee,
			data.FL,
			data.NL,
			data.PC,
			data.MT,
			data.CL,
			data.MS,
			data.SC,
			data.LN,
			data.IsFullFee,
			data.Notes,
			data.RegistrationID,
		); err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to update registration for id %s", data.RegistrationID)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to commit transaction")
		return err
	}

	return nil
}
