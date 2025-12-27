package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *rbacRepo) UpdateRole(ctx context.Context, req *entity.UpdateRoleReq) (*entity.UpdateRoleResp, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.UpdateRole")
	defer span.End()

	fnName := "repo::UpdateRole"
	var resp entity.UpdateRoleResp
	resp.ID = req.ID

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to begin transaction", fnName)
		return nil, err
	}
	defer tx.Rollback()

	query := `
		UPDATE roles
		SET
			name = ?,
			updated_at = NOW()
		WHERE id = ?
	`

	_, err = tx.ExecContext(ctx, r.db.Rebind(query), req.Name, req.ID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to update role", fnName)
		return nil, err
	}

	query = `
		DELETE FROM role_permissions
		WHERE role_id = ?
	`

	_, err = tx.ExecContext(ctx, tx.Rebind(query), req.ID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to delete role permissions", fnName)
		return nil, err
	}

	if len(req.Permissions) == 0 {
		err = tx.Commit()
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to commit transaction", fnName)
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
		_, err = tx.ExecContext(ctx, query, req.ID, v)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to insert role permissions", fnName)
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to commit transaction", fnName)
		return nil, err
	}

	return &resp, nil
}
