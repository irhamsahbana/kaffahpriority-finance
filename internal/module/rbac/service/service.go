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

func (s *rbacService) GetRoleAndPermissions(ctx context.Context, req *entity.GetRoleAndPermissionsReq) (*entity.GetRoleAndPermissionsResp, error) {
	return s.repo.GetRoleAndPermissions(ctx, req)
}

func (s *rbacService) IsHasPermission(ctx context.Context, userId, permission string) error {
	return s.repo.IsHasPermission(ctx, userId, permission)
}
