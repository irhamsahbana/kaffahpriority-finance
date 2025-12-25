package repository

import (
	"codebase-app/internal/entity"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *reportRepo) DeleteTemplate(ctx context.Context, req *entity.GetTemplateReq) error {
	fnName := "repo::DeleteTemplate"
	query := `
		UPDATE program_registration_templates
		SET deleted_at = NOW()
		WHERE id = $1
		AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, req.ID)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to delete template", fnName)
		return err
	}

	return nil
}
