package repository

import (
	"codebase-app/internal/module/report/entity"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *reportRepo) UpdateRegistrationsPaidAt(ctx context.Context, req *entity.UpdateRegisPaidAtReq) error {
	fnName := "repo::UpdateRegistrationsPaidAt"
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to begin transaction", fnName)
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE program_registrations
		SET
			paid_at = ($1::timestamp AT TIME ZONE 'Asia/Makassar'),
			allocated_at = ($2::timestamp AT TIME ZONE 'Asia/Makassar'),
			is_paid = TRUE,
			updated_at = NOW()
		WHERE id = $2
		AND deleted_at IS NULL
	`

	for _, registrationId := range req.RegistrationIds {
		if _, err := tx.ExecContext(ctx, query,
			(req.PaidAt + " 01:00:00"),
			(req.AllocatedAt + " 01:00:00"),
			registrationId); err != nil {
			log.Error().Err(err).Any("req", req).Msgf("%s - failed to update registration paid_at for id %s", fnName, registrationId)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to commit transaction", fnName)
		return err
	}

	return nil
}
