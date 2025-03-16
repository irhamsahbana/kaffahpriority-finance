package service

import (
	"codebase-app/internal/module/rbac/entity"
	"codebase-app/internal/module/rbac/ports"
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
	return s.repo.IsHasPermission(ctx, userId, permission)
}

func (s *rbacService) GetRoleAndPermissions(ctx context.Context, req *entity.GetRoleAndPermissionsReq) (*entity.GetRoleAndPermissionsResp, error) {
	return s.repo.GetRoleAndPermissions(ctx, req)
}

func (s *rbacService) CreateRole(ctx context.Context, req *entity.CreateRoleReq) (*entity.CreateRoleResp, error) {
	return s.repo.CreateRole(ctx, req)
}

func (s *rbacService) UpdateRole(ctx context.Context, req *entity.UpdateRoleReq) (*entity.UpdateRoleResp, error) {
	return s.repo.UpdateRole(ctx, req)
}

func (s *rbacService) DeleteRole(ctx context.Context, req *entity.DeleteRoleReq) error {
	return s.repo.DeleteRole(ctx, req)
}

func (s *rbacService) GetPermissions(ctx context.Context, req *entity.GetPermissionsReq) (*entity.GetPermissionsResp, error) {
	return s.repo.GetPermissions(ctx, req)
}

func (s *rbacService) UpdateRolePermissions(ctx context.Context, req *entity.UpdateRolePermissionsReq) (*entity.UpdateRolePermissionsResp, error) {
	return s.repo.UpdateRolePermissions(ctx, req)
}
