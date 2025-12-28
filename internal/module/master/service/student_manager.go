package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
)

func (s *masterService) GetStudentManagers(ctx context.Context, req *entity.GetStudentManagersReq) (*entity.GetStudentManagersResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetStudentManagers")
	defer span.End()
	return s.repo.GetStudentManagers(ctx, req)
}

func (s *masterService) CreateStudentManager(ctx context.Context, req *entity.CreateStudentManagerReq) (*entity.CreateStudentManagerResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.CreateStudentManager")
	defer span.End()
	return s.repo.CreateStudentManager(ctx, req)
}

func (s *masterService) GetStudentManager(ctx context.Context, req *entity.GetStudentManagerReq) (*entity.GetStudentManagerResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetStudentManager")
	defer span.End()
	return s.repo.GetStudentManager(ctx, req)
}

func (s *masterService) UpdateStudentManager(ctx context.Context, req *entity.UpdateStudentManagerReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.UpdateStudentManager")
	defer span.End()
	return s.repo.UpdateStudentManager(ctx, req)
}

func (s *masterService) DeleteStudentManager(ctx context.Context, req *entity.DeleteStudentManagerReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.DeleteStudentManager")
	defer span.End()
	return s.repo.DeleteStudentManager(ctx, req)
}
