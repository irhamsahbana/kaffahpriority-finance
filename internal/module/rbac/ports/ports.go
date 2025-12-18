package ports

import (
	"codebase-app/internal/entity"
	"context"
)

type RBACRepository interface {
	IsHasPermission(ctx context.Context, userId, permission string) error
	GetRoleAndPermissions(ctx context.Context, req *entity.GetRoleAndPermissionsReq) (*entity.GetRoleAndPermissionsResp, error)
	UpdateRolePermissions(ctx context.Context, req *entity.UpdateRolePermissionsReq) (*entity.UpdateRolePermissionsResp, error)
	GetRoleDetail(ctx context.Context, req *entity.GetRoleDetailReq) (*entity.GetRoleDetailResp, error)

	CreateRole(ctx context.Context, req *entity.CreateRoleReq) (*entity.CreateRoleResp, error)
	UpdateRole(ctx context.Context, req *entity.UpdateRoleReq) (*entity.UpdateRoleResp, error)
	DeleteRole(ctx context.Context, req *entity.DeleteRoleReq) error

	GetPermissions(ctx context.Context, req *entity.GetPermissionsReq) (*entity.GetPermissionsResp, error)
}

type RBACService interface {
	IsHasPermission(ctx context.Context, userId, permission string) error
	GetRoleAndPermissions(ctx context.Context, req *entity.GetRoleAndPermissionsReq) (*entity.GetRoleAndPermissionsResp, error)
	UpdateRolePermissions(ctx context.Context, req *entity.UpdateRolePermissionsReq) (*entity.UpdateRolePermissionsResp, error)
	GetRoleDetail(ctx context.Context, req *entity.GetRoleDetailReq) (*entity.GetRoleDetailResp, error)

	CreateRole(ctx context.Context, req *entity.CreateRoleReq) (*entity.CreateRoleResp, error)
	UpdateRole(ctx context.Context, req *entity.UpdateRoleReq) (*entity.UpdateRoleResp, error)
	DeleteRole(ctx context.Context, req *entity.DeleteRoleReq) error

	GetPermissions(ctx context.Context, req *entity.GetPermissionsReq) (*entity.GetPermissionsResp, error)
}
