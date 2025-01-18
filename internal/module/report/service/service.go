package service

import (
	"codebase-app/internal/module/report/entity"
	"codebase-app/internal/module/report/ports"
	"context"
)

var _ ports.ReportService = &reportService{}

type reportService struct {
	repo ports.ReportRepository
}

func NewReportService(repo ports.ReportRepository) *reportService {
	return &reportService{
		repo: repo,
	}
}

func (s *reportService) CreateTemplate(ctx context.Context, req *entity.CreateTemplateReq) (*entity.CreateTemplateResp, error) {
	return s.repo.CreateTemplate(ctx, req)
}

func (s *reportService) UpdateTemplate(ctx context.Context, req *entity.UpdateTemplateReq) (*entity.UpdateTemplateResp, error) {
	return s.repo.UpdateTemplate(ctx, req)
}

func (s *reportService) GetTemplates(ctx context.Context, req *entity.GetTemplatesReq) (*entity.GetTemplatesResp, error) {
	return s.repo.GetTemplates(ctx, req)
}

func (s *reportService) GetTemplate(ctx context.Context, req *entity.GetTemplateReq) (*entity.GetTemplateResp, error) {
	return s.repo.GetTemplate(ctx, req)
}

func (s *reportService) CreateRegistrations(ctx context.Context, req *entity.CreateRegistrationsReq) error {
	return s.repo.CreateRegistrations(ctx, req)
}
