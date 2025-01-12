package service

import (
	"codebase-app/internal/module/master/entity"
	"codebase-app/internal/module/master/ports"
	"context"
)

var _ ports.MasterService = &masterService{}

type masterService struct {
	repo ports.MasterRepository
}

func NewMasterService(repo ports.MasterRepository) *masterService {
	return &masterService{
		repo: repo,
	}
}

func (s *masterService) GetMarketers(ctx context.Context, req *entity.GetMarketersReq) (*entity.GetMarketersResp, error) {
	return s.repo.GetMarketers(ctx, req)
}

func (s *masterService) GetLecturers(ctx context.Context, req *entity.GetLecturersReq) (*entity.GetLecturersResp, error) {
	return s.repo.GetLecturers(ctx, req)
}

func (s *masterService) GetStudents(ctx context.Context, req *entity.GetStudentsReq) (*entity.GetStudentsResp, error) {
	return s.repo.GetStudents(ctx, req)
}

func (s *masterService) GetPrograms(ctx context.Context, req *entity.GetProgramsReq) (*entity.GetProgramsResp, error) {
	return s.repo.GetPrograms(ctx, req)
}
