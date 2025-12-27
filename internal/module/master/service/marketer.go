package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
)

func (s *masterService) GetMarketers(ctx context.Context, req *entity.GetMarketersReq) (*entity.GetMarketersResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetMarketers")
	defer span.End()
	return s.repo.GetMarketers(ctx, req)
}

func (s *masterService) GetMarketer(ctx context.Context, req *entity.GetMarketerReq) (*entity.GetMarketerResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetMarketer")
	defer span.End()
	return s.repo.GetMarketer(ctx, req)
}

func (s *masterService) CreateMarketer(ctx context.Context, req *entity.CreateMarketerReq) (*entity.CreateMarketerResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.CreateMarketer")
	defer span.End()
	return s.repo.CreateMarketer(ctx, req)
}

func (s *masterService) UpdateMarketer(ctx context.Context, req *entity.UpdateMarketerReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.UpdateMarketer")
	defer span.End()
	return s.repo.UpdateMarketer(ctx, req)
}

func (s *masterService) DeleteMarketer(ctx context.Context, req *entity.DeleteMarketerReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.DeleteMarketer")
	defer span.End()
	return s.repo.DeleteMarketer(ctx, req)
}
