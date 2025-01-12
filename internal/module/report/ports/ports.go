package ports

import (
	"codebase-app/internal/module/report/entity"
	"context"
)

type ReportRepository interface {
	CreateTemplate(ctx context.Context, req *entity.CreateTemplateReq) (*entity.CreateTemplateResp, error)
}

type ReportService interface {
	CreateTemplate(ctx context.Context, req *entity.CreateTemplateReq) (*entity.CreateTemplateResp, error)
}
