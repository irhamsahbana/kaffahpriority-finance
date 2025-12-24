package repository

import (
	"codebase-app/internal/entity"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *rbacRepo) DeleteRole(ctx context.Context, req *entity.DeleteRoleReq) error {
	fnName := "repo::DeleteRole"
	query := `
		UPDATE roles
		SET deleted_at = NOW()
		WHERE id = ?
		AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, r.db.Rebind(query), req.ID)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to delete role", fnName)
		return err
	}

	return nil
}
