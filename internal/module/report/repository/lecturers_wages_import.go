package repository

import (
	"codebase-app/internal/module/report/entity"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *reportRepo) ImportLecturersWages(ctx context.Context, req *entity.ImportLecturersWagesReq) error {
	var (
		fnName = "repo::ImportedLecturersWages"
	)

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to begin transaction", fnName)
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
			is_full_fee = ?,
			notes_for_lecturer_wage = ?
		WHERE
			id = ?
			AND deleted_at IS NULL
	`

	for _, data := range req.Registrations {
		if _, err := tx.ExecContext(ctx, r.db.Rebind(query),
			data.ProgramMeetings,
			data.InitialFee,
			data.FL,
			data.NL,
			data.IsFullFee,
			data.Notes,
			data.RegistrationId,
		); err != nil {
			log.Error().Err(err).Any("req", req).Msgf("%s - failed to update registration for id %s", fnName, data.RegistrationId)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to commit transaction", fnName)
		return err
	}

	return nil
}
