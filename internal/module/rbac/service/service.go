package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	ports "codebase-app/internal/ports/module/rbac"
	"context"
)

var _ ports.RBACService = &rbacService{}

type rbacService struct {
	repo ports.RBACRepository
}

func NewRBACService(repo ports.RBACRepository) *rbacService {
	return &rbacService{
		repo: repo,
	}
}

func (s *rbacService) IsHasPermission(ctx context.Context, userId, permission string) error {
	ctx, span := tracing.StartSpan(ctx, "service.IsHasPermission")
	defer span.End()
	return s.repo.IsHasPermission(ctx, userId, permission)
}

func (s *rbacService) GetRoleAndPermissions(ctx context.Context, req *entity.GetRoleAndPermissionsReq) (*entity.GetRoleAndPermissionsResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetRoleAndPermissions")
	defer span.End()
	return s.repo.GetRoleAndPermissions(ctx, req)
}

func (s *rbacService) GetRoleDetail(ctx context.Context, req *entity.GetRoleDetailReq) (*entity.GetRoleDetailResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetRoleDetail")
	defer span.End()
	return s.repo.GetRoleDetail(ctx, req)
}

func (s *rbacService) CreateRole(ctx context.Context, req *entity.CreateRoleReq) (*entity.CreateRoleResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.CreateRole")
	defer span.End()
	return s.repo.CreateRole(ctx, req)
}

func (s *rbacService) UpdateRole(ctx context.Context, req *entity.UpdateRoleReq) (*entity.UpdateRoleResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.UpdateRole")
	defer span.End()
	return s.repo.UpdateRole(ctx, req)
}

func (s *rbacService) DeleteRole(ctx context.Context, req *entity.DeleteRoleReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.DeleteRole")
	defer span.End()
	return s.repo.DeleteRole(ctx, req)
}

func (s *rbacService) GetPermissions(ctx context.Context, req *entity.GetPermissionsReq) (*entity.GetPermissionsResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetPermissions")
	defer span.End()
	return s.repo.GetPermissions(ctx, req)
}

func (s *rbacService) UpdateRolePermissions(ctx context.Context, req *entity.UpdateRolePermissionsReq) (*entity.UpdateRolePermissionsResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.UpdateRolePermissions")
	defer span.End()
	return s.repo.UpdateRolePermissions(ctx, req)
}
