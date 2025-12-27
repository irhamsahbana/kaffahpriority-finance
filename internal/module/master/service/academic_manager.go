package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
)

func (s *masterService) GetAcademicManagers(ctx context.Context, req *entity.GetAcademicManagersReq) (*entity.GetAcademicManagersResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetAcademicManagers")
	defer span.End()
	return s.repo.GetAcademicManagers(ctx, req)
}

func (s *masterService) CreateAcademicManager(ctx context.Context, req *entity.CreateAcademicManagerReq) (*entity.CreateAcademicManagerResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.CreateAcademicManager")
	defer span.End()
	return s.repo.CreateAcademicManager(ctx, req)
}

func (s *masterService) GetAcademicManager(ctx context.Context, req *entity.GetAcademicManagerReq) (*entity.GetAcademicManagerResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetAcademicManager")
	defer span.End()
	return s.repo.GetAcademicManager(ctx, req)
}

func (s *masterService) UpdateAcademicManager(ctx context.Context, req *entity.UpdateAcademicManagerReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.UpdateAcademicManager")
	defer span.End()
	return s.repo.UpdateAcademicManager(ctx, req)
}

func (s *masterService) DeleteAcademicManager(ctx context.Context, req *entity.DeleteAcademicManagerReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.DeleteAcademicManager")
	defer span.End()
	return s.repo.DeleteAcademicManager(ctx, req)
}
