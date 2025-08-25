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
				AND (p.name = ? OR p.name = 'all_access')
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
		resp.Items = append(resp.Items, *v)
	}

	return &resp, nil
}

func (r *rbacRepo) CreateRole(ctx context.Context, req *entity.CreateRoleReq) (*entity.CreateRoleResp, error) {
	var resp entity.CreateRoleResp
	resp.ID = ulid.Make().String()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::CreateRole - failed to begin transaction")
		return nil, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO roles (id, name)
		VALUES (?, ?)
	`

	_, err = tx.ExecContext(ctx, r.db.Rebind(query), resp.ID, req.Name)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::CreateRole - failed to create role")
		return nil, err
	}

	if len(req.Permissions) == 0 {
		err = tx.Commit()
		if err != nil {
			log.Error().Err(err).Any("req", req).Msg("repo::CreateRole - failed to commit transaction")
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
			log.Error().Err(err).Any("req", req).Msg("repo::CreateRole - failed to insert role permissions")
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::CreateRole - failed to commit transaction")
		return nil, err
	}

	return &resp, nil
}

func (r *rbacRepo) UpdateRole(ctx context.Context, req *entity.UpdateRoleReq) (*entity.UpdateRoleResp, error) {
	var resp entity.UpdateRoleResp
	resp.ID = req.ID

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateRole - failed to begin transaction")
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
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateRole - failed to update role")
		return nil, err
	}

	query = `
		DELETE FROM role_permissions
		WHERE role_id = ?
	`

	_, err = tx.ExecContext(ctx, tx.Rebind(query), req.ID)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateRole - failed to delete role permissions")
		return nil, err
	}

	if len(req.Permissions) == 0 {
		err = tx.Commit()
		if err != nil {
			log.Error().Err(err).Any("req", req).Msg("repo::UpdateRole - failed to commit transaction")
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
			log.Error().Err(err).Any("req", req).Msg("repo::UpdateRole - failed to insert role permissions")
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateRole - failed to commit transaction")
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

	_, err := r.db.ExecContext(ctx, r.db.Rebind(query), req.ID)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::DeleteRole - failed to delete role")
		return err
	}

	return nil
}

func (r *rbacRepo) GetPermissions(ctx context.Context, req *entity.GetPermissionsReq) (*entity.GetPermissionsResp, error) {
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
		log.Error().Err(err).Any("req", req).Msg("repo::GetPermissions - failed to get permissions")
		return nil, err
	}

	return &resp, nil
}

func (r *rbacRepo) UpdateRolePermissions(ctx context.Context, req *entity.UpdateRolePermissionsReq) (*entity.UpdateRolePermissionsResp, error) {
	var resp entity.UpdateRolePermissionsResp
	resp.ID = req.RoleID

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateRolePermissions - failed to begin transaction")
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
			log.Error().Err(err).Any("req", req).Msg("repo::UpdateRolePermissions - failed to get role that has all_access permission")
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
			log.Error().Err(err).Any("req", req).Msg("repo::UpdateRolePermissions - failed to get role id")
			return nil, err
		}

		if roleID == req.RoleID && roleThatHasAllAccess == 1 {
			return nil, errmsg.NewCustomErrors(400).SetMessage("Role ini adalah role terakhir yang memiliki all_access permission!")
		}

		if roleThatHasAllAccess == 1 {
			return nil, errmsg.NewCustomErrors(400).SetMessage("Role ini adalah role terakhir yang memiliki all_access permission!")
		}
	}

	query := `
		DELETE FROM role_permissions
		WHERE role_id = ?
	`

	_, err = tx.ExecContext(ctx, tx.Rebind(query), req.RoleID)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateRolePermissions - failed to delete role permissions")
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
			log.Error().Err(err).Any("req", req).Msg("repo::UpdateRolePermissions - failed to insert role permissions")
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateRolePermissions - failed to commit transaction")
		return nil, err
	}

	return &resp, nil
}

func (r *rbacRepo) GetRoleDetail(ctx context.Context, req *entity.GetRoleDetailReq) (*entity.GetRoleDetailResp, error) {
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
		log.Error().Err(err).Any("req", req).Msg("repo::GetRoleDetail - failed to get role detail")
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
