package ports

import (
	"codebase-app/internal/entity"
	"context"
)

type ActivityLogRepository interface {
	CreateActivityLog(ctx context.Context, req *entity.ActivityLog) error
	GetActivityLogs(ctx context.Context, req *entity.GetActivityLogsReq) (*entity.GetActivityLogsResp, error)
}

type ActivityLogService interface {
	CreateActivityLog(ctx context.Context, req *entity.ActivityLog) error
	GetActivityLogs(ctx context.Context, req *entity.GetActivityLogsReq) (*entity.GetActivityLogsResp, error)
}
