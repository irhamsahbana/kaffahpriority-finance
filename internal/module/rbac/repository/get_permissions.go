package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *rbacRepo) GetPermissions(ctx context.Context, req *entity.GetPermissionsReq) (*entity.GetPermissionsResp, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetPermissions")
	defer span.End()

	fnName := "repo::GetPermissions"
	var (
		resp entity.GetPermissionsResp
	)
	resp.Items = make([]entity.Permission, 0)

	query := `
		SELECT
			id,
			name,
			description
		FROM permissions
		WHERE deleted_at IS NULL
		ORDER BY id ASC
	`

	err := r.db.SelectContext(ctx, &resp.Items, query)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to get permissions", fnName)
		return nil, err
	}

	return &resp, nil
}
