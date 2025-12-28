package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
)

func (s *masterService) CreateProgram(ctx context.Context, req *entity.CreateProgramReq) (*entity.CreateProgramResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.CreateProgram")
	defer span.End()
	return s.repo.CreateProgram(ctx, req)
}

func (s *masterService) GetPrograms(ctx context.Context, req *entity.GetProgramsReq) (*entity.GetProgramsResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetPrograms")
	defer span.End()
	return s.repo.GetPrograms(ctx, req)
}

func (s *masterService) GetProgram(ctx context.Context, req *entity.GetProgramReq) (*entity.GetProgramResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetProgram")
	defer span.End()
	return s.repo.GetProgram(ctx, req)
}

func (s *masterService) UpdateProgram(ctx context.Context, req *entity.UpdateProgramReq) (*entity.UpdateProgramResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.UpdateProgram")
	defer span.End()
	return s.repo.UpdateProgram(ctx, req)
}

func (s *masterService) DeleteProgram(ctx context.Context, req *entity.DeleteProgramReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.DeleteProgram")
	defer span.End()
	return s.repo.DeleteProgram(ctx, req)
}
