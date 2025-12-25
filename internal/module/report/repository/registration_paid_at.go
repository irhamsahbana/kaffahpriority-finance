package repository

import (
	"codebase-app/internal/entity"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *reportRepo) UpdateRegistrationsPaidAt(ctx context.Context, req *entity.UpdateRegisPaidAtReq) error {
	fnName := "repo::UpdateRegistrationsPaidAt"
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to begin transaction", fnName)
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
		WHERE id = $3
		AND deleted_at IS NULL
	`

	for _, registrationId := range req.RegistrationIds {
		if _, err := tx.ExecContext(ctx, query,
			req.PaidAt+" "+req.PaidAtTime,
			req.AllocatedAt+" "+req.AllocatedAtTime,
			registrationId,
		); err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to update registration paid_at for id %s", fnName, registrationId)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to commit transaction", fnName)
		return err
	}

	return nil
}
