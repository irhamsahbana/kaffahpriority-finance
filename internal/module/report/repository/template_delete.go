package repository

import (
	"codebase-app/internal/module/report/entity"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *reportRepo) DeleteTemplate(ctx context.Context, req *entity.GetTemplateReq) error {
	query := `
		UPDATE program_registration_templates
		SET deleted_at = NOW()
		WHERE id = $1
		AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, req.ID)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repository::DeleteTemplate - error deleting template")
		return err
	}

	return nil
}
