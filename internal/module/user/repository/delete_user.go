package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *userRepo) DeleteUser(ctx context.Context, req *entity.DeleteUserReq) error {
	ctx, span := tracing.StartSpan(ctx, "repo.DeleteUser")
	defer span.End()

	fnName := "repo::DeleteUser"
	query := `
		UPDATE
			users
		SET
			deleted_at = NOW()
		WHERE
			id = ?
		AND
			deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, r.db.Rebind(query), req.ID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to delete user", fnName)
		return err
	}

	return nil
}
