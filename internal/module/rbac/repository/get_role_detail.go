package repository

import (
	"codebase-app/internal/entity"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *rbacRepo) GetRoleDetail(ctx context.Context, req *entity.GetRoleDetailReq) (*entity.GetRoleDetailResp, error) {
	fnName := "repo::GetRoleDetail"
	var (
		resp entity.GetRoleDetailResp
	)
	resp.Permissions = make([]entity.Permission, 0)

	type RolePermission struct {
		RoleID       string  `db:"role_id"`
		Role         string  `db:"role"`
		PermissionId *string `db:"permission_id"`
		Permission   *string `db:"permission"`
		Description  *string `db:"description"`
	}

	query := `
		SELECT
			r.id AS role_id,
			r.name AS role,
			p.id AS permission_id,
			p.name AS permission,
			p.description
		FROM roles r
		LEFT JOIN role_permissions rp ON r.id = rp.role_id
		LEFT JOIN permissions p ON rp.permission_id = p.id
		WHERE r.id = ?
		AND r.deleted_at IS NULL
		ORDER BY r.id ASC
	`

	data := make([]RolePermission, 0)
	err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), req.RoleID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to get role detail", fnName)
		return nil, err
	}

	rolePermissionMap := make(map[string]*entity.RolePermissionItem)
	for _, v := range data {
		if _, ok := rolePermissionMap[v.RoleID]; !ok {
			rolePermissionMap[v.RoleID] = &entity.RolePermissionItem{
				RoleID:      v.RoleID,
				Role:        v.Role,
				Permissions: make([]entity.Permission, 0),
			}
		}

		if v.Permission != nil {
			rolePermissionMap[v.RoleID].Permissions = append(rolePermissionMap[v.RoleID].Permissions, entity.Permission{
				ID:          *v.PermissionId,
				Name:        *v.Permission,
				Description: *v.Description,
			})
		}
	}

	for _, v := range rolePermissionMap {
		resp.RolePermissionItem = *v
	}

	return &resp, nil
}
