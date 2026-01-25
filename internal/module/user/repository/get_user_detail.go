package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/pkg/errmsg"
	"context"
	"database/sql"

	"github.com/rs/zerolog/log"
)

func (r *userRepo) GetUser(ctx context.Context, req *entity.GetUserReq) (*entity.GetUserResp, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetUser")
	defer span.End()

	var resp entity.GetUserResp

	query := `
		SELECT
			u.id,
			u.role_id,
			u.name,
			u.email
		FROM
			users u
		WHERE
			u.id = ?
		AND
			u.deleted_at IS NULL
	`

	err := r.db.GetContext(ctx, &resp, r.db.Rebind(query), req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Ctx(ctx).Warn().Err(err).Any("req", req).Msg("user not found")
			return nil, errmsg.NewCustomErrors(404).SetMessage("Data tidak ditemukan")
		}
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msg("failed to fetch user")
		return nil, err
	}

	return &resp, nil
}
