package service

import (
	"codebase-app/internal/entity"
	"context"
)

func (s *masterService) GetAcademicManagers(ctx context.Context, req *entity.GetAcademicManagersReq) (*entity.GetAcademicManagersResp, error) {
	return s.repo.GetAcademicManagers(ctx, req)
}

func (s *masterService) CreateAcademicManager(ctx context.Context, req *entity.CreateAcademicManagerReq) (*entity.CreateAcademicManagerResp, error) {
	return s.repo.CreateAcademicManager(ctx, req)
}

func (s *masterService) GetAcademicManager(ctx context.Context, req *entity.GetAcademicManagerReq) (*entity.GetAcademicManagerResp, error) {
	return s.repo.GetAcademicManager(ctx, req)
}

func (s *masterService) UpdateAcademicManager(ctx context.Context, req *entity.UpdateAcademicManagerReq) error {
	return s.repo.UpdateAcademicManager(ctx, req)
}

func (s *masterService) DeleteAcademicManager(ctx context.Context, req *entity.DeleteAcademicManagerReq) error {
	return s.repo.DeleteAcademicManager(ctx, req)
}
