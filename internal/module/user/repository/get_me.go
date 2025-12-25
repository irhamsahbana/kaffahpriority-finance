package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/pkg/errmsg"
	"context"
	"database/sql"

	"github.com/rs/zerolog/log"
)

func (r *userRepo) GetMe(ctx context.Context, req *entity.GetMeReq) (*entity.GetMeResp, error) {
	fnName := "repo::GetMe"
	var (
		resp entity.GetMeResp
	)
	resp.Permissions = make([]entity.Permission, 0)

	query := `
		SELECT
			u.id,
			u.role_id,
			r.name as role,
			u.name,
			u.email
		FROM
			users u
		JOIN
			roles r ON r.id = u.role_id
		WHERE
			u.id = ?
		AND
			u.deleted_at IS NULL
	`
	err := r.db.GetContext(ctx, &resp, r.db.Rebind(query), req.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Ctx(ctx).Warn().Err(err).Any("req", req).Msgf("%s - user not found", fnName)
			return nil, errmsg.NewCustomErrors(404).SetMessage("Data tidak ditemukan")
		}
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to fetch user", fnName)
		return nil, err
	}

	query = `
		SELECT
			p.id,
			p.name,
			p.description
		FROM
			permissions p
		JOIN
			role_permissions rp ON rp.permission_id = p.id
		JOIN
			roles r ON r.id = rp.role_id
		JOIN
			users u ON u.role_id = r.id
		WHERE
			u.id = ?
			AND u.deleted_at IS NULL
			AND p.deleted_at IS NULL
	`

	err = r.db.SelectContext(ctx, &resp.Permissions, r.db.Rebind(query), req.UserID)
	if err != nil && err != sql.ErrNoRows {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to fetch permissions", fnName)
		return nil, err
	}

	if len(resp.Permissions) == 0 {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msgf("%s - permissions not found", fnName)
	}

	return &resp, nil
}
