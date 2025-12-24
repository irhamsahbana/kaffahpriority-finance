package repository

import (
	"codebase-app/internal/entity"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *rbacRepo) GetPermissions(ctx context.Context, req *entity.GetPermissionsReq) (*entity.GetPermissionsResp, error) {
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
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to get permissions", fnName)
		return nil, err
	}

	return &resp, nil
}
