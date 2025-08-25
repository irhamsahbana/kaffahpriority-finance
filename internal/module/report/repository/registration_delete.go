package repository

import (
	"codebase-app/internal/module/report/entity"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *reportRepo) DeleteRegistration(ctx context.Context, req *entity.GetRegistrationReq) error {
	query := `
		UPDATE program_registrations
		SET deleted_at = NOW()
		WHERE id = $1
		AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, req.ID)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::DeleteRegistration - error deleting registration")
		return err
	}

	return nil
}
