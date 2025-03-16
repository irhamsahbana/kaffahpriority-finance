package repository

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/module/rbac/entity"
	"codebase-app/internal/module/rbac/ports"
	"codebase-app/pkg/errmsg"
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

var _ ports.RBACRepository = &rbacRepo{}

type rbacRepo struct {
	db *sqlx.DB
}

func NewRBACRepository() *rbacRepo {
	return &rbacRepo{
		db: adapter.Adapters.Postgres,
	}
}

func (r *rbacRepo) IsHasPermission(ctx context.Context, userId, permission string) error {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM users u
			JOIN roles r ON u.role_id = r.id
			JOIN role_permissions rp ON r.id = rp.role_id
			JOIN permissions p ON rp.permission_id = p.id
			WHERE
				u.id = ?
				AND p.name = ?
		)
	`

	var isHasPermission bool
	err := r.db.GetContext(ctx, &isHasPermission, query, userId, permission)
	if err != nil {
		log.Error().Err(err).Msg("repo::IsHasPermission - failed to get permission")
		return err
	}

	if !isHasPermission {
		log.Warn().Msg("repo::IsHasPermission - user does not have permission")
		return errmsg.NewCustomErrors(403).SetMessage("Anda tidak memiliki hak akses" + permission + "!")
	}

	return nil
}

func (r *rbacRepo) GetRoleAndPermissions(ctx context.Context, req *entity.GetRoleAndPermissionsReq) (*entity.GetRoleAndPermissionsResp, error) {
	var (
		resp entity.GetRoleAndPermissionsResp
	)
	resp.Items = make([]entity.RolePermissionItem, 0)

	type RolePermission struct {
		RoleId       string  `db:"role_id"`
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
	`

	data := make([]RolePermission, 0)
	err := r.db.SelectContext(ctx, &data, query)
	if err != nil {
		log.Error().Err(err).Msg("repo::GetRoleAndPermissions - failed to get role and permissions")
		return nil, err
	}

	rolePermissionMap := make(map[string]*entity.RolePermissionItem)
	for _, v := range data {
		if _, ok := rolePermissionMap[v.RoleId]; !ok {
			rolePermissionMap[v.RoleId] = &entity.RolePermissionItem{
				RoleId:      v.RoleId,
				Role:        v.Role,
				Permissions: make([]entity.Permission, 0),
			}
		}

		if v.Permission != nil {
			rolePermissionMap[v.RoleId].Permissions = append(rolePermissionMap[v.RoleId].Permissions, entity.Permission{
				Id:          *v.PermissionId,
				Name:        *v.Permission,
				Description: *v.Description,
			})
		}
	}

	for _, v := range rolePermissionMap {
		resp.Items = append(resp.Items, *v)
	}

	return &resp, nil
}
