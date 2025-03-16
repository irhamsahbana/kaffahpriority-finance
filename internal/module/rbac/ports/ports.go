package ports

import (
	"codebase-app/internal/module/rbac/entity"
	"context"
)

type RBACRepository interface {
	IsHasPermission(ctx context.Context, userId, permission string) error
	GetRoleAndPermissions(ctx context.Context, req *entity.GetRoleAndPermissionsReq) (*entity.GetRoleAndPermissionsResp, error)
}

type RBACService interface {
	IsHasPermission(ctx context.Context, userId, permission string) error
	GetRoleAndPermissions(ctx context.Context, req *entity.GetRoleAndPermissionsReq) (*entity.GetRoleAndPermissionsResp, error)
}
