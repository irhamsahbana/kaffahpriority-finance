package service

import (
	"codebase-app/internal/module/master/entity"
	"context"
)

func (s *masterService) CreateProgram(ctx context.Context, req *entity.CreateProgramReq) (*entity.CreateProgramResp, error) {
	return s.repo.CreateProgram(ctx, req)
}

func (s *masterService) GetPrograms(ctx context.Context, req *entity.GetProgramsReq) (*entity.GetProgramsResp, error) {
	return s.repo.GetPrograms(ctx, req)
}

func (s *masterService) GetProgram(ctx context.Context, req *entity.GetProgramReq) (*entity.GetProgramResp, error) {
	return s.repo.GetProgram(ctx, req)
}

func (s *masterService) UpdateProgram(ctx context.Context, req *entity.UpdateProgramReq) (*entity.UpdateProgramResp, error) {
	return s.repo.UpdateProgram(ctx, req)
}

func (s *masterService) DeleteProgram(ctx context.Context, req *entity.DeleteProgramReq) error {
	return s.repo.DeleteProgram(ctx, req)
}
