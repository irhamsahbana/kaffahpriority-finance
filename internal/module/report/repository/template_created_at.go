package repository

import (
	"codebase-app/internal/infrastructure/tracing"
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

func (r *reportRepo) GetTemplateCreatedAt(ctx context.Context, id string) (time.Time, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetTemplateCreatedAt")
	defer span.End()

	var createdAt time.Time
	query := `SELECT created_at FROM program_registration_templates WHERE id = ? AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &createdAt, r.db.Rebind(query), id)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("id", id).Msg("repo::GetTemplateCreatedAt - failed to get template created_at")
		return time.Time{}, err
	}

	return createdAt, nil
}

func (r *reportRepo) UpdateTemplateCreatedAt(ctx context.Context, id string, createdAt time.Time) error {
	ctx, span := tracing.StartSpan(ctx, "repo.UpdateTemplateCreatedAt")
	defer span.End()

	query := `UPDATE program_registration_templates SET created_at = ? WHERE id = ?`

	_, err := r.db.ExecContext(ctx, r.db.Rebind(query), createdAt, id)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("id", id).Time("created_at", createdAt).Msg("repo::UpdateTemplateCreatedAt - failed to update template created_at")
		return err
	}

	return nil
}
