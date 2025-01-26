package service

import (
	"codebase-app/internal/module/master/entity"
	"context"
)

func (s *masterService) GetStudentManagers(ctx context.Context, req *entity.GetStudentManagersReq) (*entity.GetStudentManagersResp, error) {
	return s.repo.GetStudentManagers(ctx, req)
}

func (s *masterService) CreateStudentManager(ctx context.Context, req *entity.CreateStudentManagerReq) (*entity.CreateStudentManagerResp, error) {
	return s.repo.CreateStudentManager(ctx, req)
}

func (s *masterService) GetStudentManager(ctx context.Context, req *entity.GetStudentManagerReq) (*entity.GetStudentManagerResp, error) {
	return s.repo.GetStudentManager(ctx, req)
}

func (s *masterService) UpdateStudentManager(ctx context.Context, req *entity.UpdateStudentManagerReq) error {
	return s.repo.UpdateStudentManager(ctx, req)
}

func (s *masterService) DeleteStudentManager(ctx context.Context, req *entity.DeleteStudentManagerReq) error {
	return s.repo.DeleteStudentManager(ctx, req)
}
