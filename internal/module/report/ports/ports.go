package ports

import (
	"codebase-app/internal/module/report/entity"
	"context"
)

type ReportRepository interface {
	GetTemplates(ctx context.Context, req *entity.GetTemplatesReq) (*entity.GetTemplatesResp, error)
	GetTemplate(ctx context.Context, req *entity.GetTemplateReq) (*entity.GetTemplateResp, error)
	CreateTemplate(ctx context.Context, req *entity.CreateTemplateReq) (*entity.CreateTemplateResp, error)
	UpdateTemplate(ctx context.Context, req *entity.UpdateTemplateReq) (*entity.UpdateTemplateResp, error)
}

type ReportService interface {
	GetTemplates(ctx context.Context, req *entity.GetTemplatesReq) (*entity.GetTemplatesResp, error)
	GetTemplate(ctx context.Context, req *entity.GetTemplateReq) (*entity.GetTemplateResp, error)
	CreateTemplate(ctx context.Context, req *entity.CreateTemplateReq) (*entity.CreateTemplateResp, error)
	UpdateTemplate(ctx context.Context, req *entity.UpdateTemplateReq) (*entity.UpdateTemplateResp, error)
}
