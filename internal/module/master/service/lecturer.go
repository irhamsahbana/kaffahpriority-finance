package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
)

func (s *masterService) GetLecturers(ctx context.Context, req *entity.GetLecturersReq) (*entity.GetLecturersResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetLecturers")
	defer span.End()
	return s.repo.GetLecturers(ctx, req)
}

func (s *masterService) GetLecturer(ctx context.Context, req *entity.GetLecturerReq) (*entity.GetLecturerResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetLecturer")
	defer span.End()
	return s.repo.GetLecturer(ctx, req)
}

func (s *masterService) CreateLecturer(ctx context.Context, req *entity.CreateLecturerReq) (*entity.CreateLecturerResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.CreateLecturer")
	defer span.End()
	return s.repo.CreateLecturer(ctx, req)
}

func (s *masterService) UpdateLecturer(ctx context.Context, req *entity.UpdateLecturerReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.UpdateLecturer")
	defer span.End()
	return s.repo.UpdateLecturer(ctx, req)
}

func (s *masterService) DeleteLecturer(ctx context.Context, req *entity.DeleteLecturerReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.DeleteLecturer")
	defer span.End()
	return s.repo.DeleteLecturer(ctx, req)
}
