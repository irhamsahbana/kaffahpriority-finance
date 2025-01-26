package service

import (
	"codebase-app/internal/module/master/entity"
	"context"
)

func (s *masterService) GetMarketers(ctx context.Context, req *entity.GetMarketersReq) (*entity.GetMarketersResp, error) {
	return s.repo.GetMarketers(ctx, req)
}

func (s *masterService) GetMarketer(ctx context.Context, req *entity.GetMarketerReq) (*entity.GetMarketerResp, error) {
	return s.repo.GetMarketer(ctx, req)
}

func (s *masterService) CreateMarketer(ctx context.Context, req *entity.CreateMarketerReq) (*entity.CreateMarketerResp, error) {
	return s.repo.CreateMarketer(ctx, req)
}

func (s *masterService) UpdateMarketer(ctx context.Context, req *entity.UpdateMarketerReq) error {
	return s.repo.UpdateMarketer(ctx, req)
}

func (s *masterService) DeleteMarketer(ctx context.Context, req *entity.DeleteMarketerReq) error {
	return s.repo.DeleteMarketer(ctx, req)
}
