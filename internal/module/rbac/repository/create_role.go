package repository

import (
	"codebase-app/internal/entity"
	"context"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (r *rbacRepo) CreateRole(ctx context.Context, req *entity.CreateRoleReq) (*entity.CreateRoleResp, error) {
	fnName := "repo::CreateRole"
	var resp entity.CreateRoleResp
	resp.ID = ulid.Make().String()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to begin transaction", fnName)
		return nil, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO roles (id, name)
		VALUES (?, ?)
	`

	_, err = tx.ExecContext(ctx, r.db.Rebind(query), resp.ID, req.Name)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to create role", fnName)
		return nil, err
	}

	if len(req.Permissions) == 0 {
		err = tx.Commit()
		if err != nil {
			log.Error().Err(err).Any("req", req).Msgf("%s - failed to commit transaction", fnName)
			return nil, err
		}

		return &resp, nil
	}

	query = `
		INSERT INTO role_permissions (role_id, permission_id)
		VALUES (?, ?)
	`
	query = tx.Rebind(query)

	for _, v := range req.Permissions {
		_, err = tx.ExecContext(ctx, query, resp.ID, v)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msgf("%s - failed to insert role permissions", fnName)
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to commit transaction", fnName)
		return nil, err
	}

	return &resp, nil
}
