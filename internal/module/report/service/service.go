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
