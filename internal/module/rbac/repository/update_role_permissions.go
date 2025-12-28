package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/pkg/errmsg"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *rbacRepo) UpdateRolePermissions(ctx context.Context, req *entity.UpdateRolePermissionsReq) (*entity.UpdateRolePermissionsResp, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.UpdateRolePermissions")
	defer span.End()

	fnName := "repo::UpdateRolePermissions"
	var resp entity.UpdateRolePermissionsResp
	resp.ID = req.RoleID

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to begin transaction", fnName)
		return nil, err
	}
	defer tx.Rollback()

	if len(req.Permissions) == 0 {
		// check if this role is the last role that has all_access permission
		roleThatHasAllAccess := 0
		query := `
			SELECT COUNT(*)
			FROM role_permissions rp
			JOIN permissions p ON rp.permission_id = p.id
			WHERE p.name = 'all_access'
		`

		err = tx.GetContext(ctx, &roleThatHasAllAccess, tx.Rebind(query))
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to get role that has all_access permission", fnName)
			return nil, err
		}

		query = `
			SELECT r.id
			FROM users u
			JOIN roles r ON u.role_id = r.id
			WHERE u.id = ?
		`

		var roleID string
		err = tx.GetContext(ctx, &roleID, tx.Rebind(query), req.UserID)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to get role id", fnName)
			return nil, err
		}

		if roleID == req.RoleID && roleThatHasAllAccess == 1 {
			log.Ctx(ctx).Warn().Msgf("%s - role %s is the last role that has all_access permission", fnName, req.RoleID)
			return nil, errmsg.NewCustomErrors(400).SetMessage("Role ini adalah role terakhir yang memiliki all_access permission!")
		}

		if roleThatHasAllAccess == 1 {
			log.Ctx(ctx).Warn().Msgf("%s - role %s is the last role that has all_access permission", fnName, req.RoleID)
			return nil, errmsg.NewCustomErrors(400).SetMessage("Role ini adalah role terakhir yang memiliki all_access permission!")
		}
	}

	query := `
		DELETE FROM role_permissions
		WHERE role_id = ?
	`

	_, err = tx.ExecContext(ctx, tx.Rebind(query), req.RoleID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to delete role permissions", fnName)
		return nil, err
	}

	query = `
		INSERT INTO role_permissions (role_id, permission_id)
		VALUES (?, ?)
	`
	query = tx.Rebind(query)

	for _, v := range req.Permissions {
		_, err = tx.ExecContext(ctx, query, req.RoleID, v)
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
