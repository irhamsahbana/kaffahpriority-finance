package repository

import (
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/pkg/errmsg"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *rbacRepo) IsHasPermission(ctx context.Context, userId, permission string) error {
	ctx, span := tracing.StartSpan(ctx, "repo.IsHasPermission")
	defer span.End()

	fnName := "repo::IsHasPermission"
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM users u
			JOIN roles r ON u.role_id = r.id
			JOIN role_permissions rp ON r.id = rp.role_id
			JOIN permissions p ON rp.permission_id = p.id
			WHERE
				u.id = ?
				AND (p.name = ? OR p.name = 'all_access')
		)
	`

	var isHasPermission bool
	err := r.db.GetContext(ctx, &isHasPermission, query, userId, permission)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("%s - failed to get permission", fnName)
		return err
	}

	if !isHasPermission {
		log.Ctx(ctx).Warn().Msgf("%s - user does not have permission", fnName)
		return errmsg.NewCustomErrors(403).SetMessage("Anda tidak memiliki hak akses" + permission + "!")
	}

	return nil
}
