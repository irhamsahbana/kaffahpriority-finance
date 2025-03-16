package repository

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/module/rbac/entity"
	"codebase-app/internal/module/rbac/ports"
	"codebase-app/pkg/errmsg"
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/oklog/ulid/v2"
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
		WHERE r.deleted_at IS NULL
		ORDER BY r.id ASC
	`

	data := make([]RolePermission, 0)
	err := r.db.SelectContext(ctx, &data, query)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetRoleAndPermissions - failed to get role and permissions")
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

func (r *rbacRepo) CreateRole(ctx context.Context, req *entity.CreateRoleReq) (*entity.CreateRoleResp, error) {
	var resp entity.CreateRoleResp
	resp.Id = ulid.Make().String()
	query := `
		INSERT INTO roles (id, name)
		VALUES (?, ?)
	`

	_, err := r.db.ExecContext(ctx, r.db.Rebind(query), resp.Id, req.Name)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::CreateRole - failed to create role")
		return nil, err
	}

	return &resp, nil
}

func (r *rbacRepo) UpdateRole(ctx context.Context, req *entity.UpdateRoleReq) (*entity.UpdateRoleResp, error) {
	var resp entity.UpdateRoleResp
	resp.Id = req.Id
	query := `
		UPDATE roles
		SET
			name = ?,
			updated_at = NOW()
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, r.db.Rebind(query), req.Name, req.Id)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateRole - failed to update role")
		return nil, err
	}

	return &resp, nil
}

func (r *rbacRepo) DeleteRole(ctx context.Context, req *entity.DeleteRoleReq) error {
	query := `
		UPDATE roles
		SET deleted_at = NOW()
		WHERE id = ?
		AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, r.db.Rebind(query), req.Id)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::DeleteRole - failed to delete role")
		return err
	}

	return nil
}
