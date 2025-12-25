package repository

import (
	"codebase-app/internal/entity"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *reportRepo) DeleteRegistration(ctx context.Context, req *entity.GetRegistrationReq) error {
	fnName := "repo::DeleteRegistration"
	query := `
		UPDATE program_registrations
		SET deleted_at = NOW()
		WHERE id = $1
		AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, req.ID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to delete registration", fnName)
		return err
	}

	return nil
}
