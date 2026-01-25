package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *userRepo) GetUsers(ctx context.Context, req *entity.GetUsersReq) (*entity.GetUsersResp, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetUsers")
	defer span.End()

	type dao struct {
		TotalData int `db:"total_data"`
		entity.UserItem
	}
	var (
		resp = new(entity.GetUsersResp)
		data = make([]dao, 0)
		args = make([]any, 0)
	)
	resp.Items = make([]entity.UserItem, 0)

	query := `
		SELECT
			COUNT(*) OVER() AS total_data,
			u.id,
			r.id as role_id,
			u.name,
			r.name as role,
			u.email
		FROM
			users u
		JOIN
			roles r ON r.id = u.role_id
		WHERE
			u.deleted_at IS NULL
		LIMIT ? OFFSET ?
	`

	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msg("failed to fetch data")
		return nil, err
	}

	for _, item := range data {
		resp.Meta.TotalData = item.TotalData
		resp.Items = append(resp.Items, item.UserItem)
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	return resp, nil
}
