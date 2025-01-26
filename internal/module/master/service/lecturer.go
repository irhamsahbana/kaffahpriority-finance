package service

import (
	"codebase-app/internal/module/master/entity"
	"context"
)

func (s *masterService) GetLecturers(ctx context.Context, req *entity.GetLecturersReq) (*entity.GetLecturersResp, error) {
	return s.repo.GetLecturers(ctx, req)
}

func (s *masterService) GetLecturer(ctx context.Context, req *entity.GetLecturerReq) (*entity.GetLecturerResp, error) {
	return s.repo.GetLecturer(ctx, req)
}

func (s *masterService) CreateLecturer(ctx context.Context, req *entity.CreateLecturerReq) (*entity.CreateLecturerResp, error) {
	return s.repo.CreateLecturer(ctx, req)
}

func (s *masterService) UpdateLecturer(ctx context.Context, req *entity.UpdateLecturerReq) error {
	return s.repo.UpdateLecturer(ctx, req)
}

func (s *masterService) DeleteLecturer(ctx context.Context, req *entity.DeleteLecturerReq) error {
	return s.repo.DeleteLecturer(ctx, req)
}
