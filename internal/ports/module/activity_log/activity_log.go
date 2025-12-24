package ports

import (
	"codebase-app/internal/entity"
	"context"
)

type ActivityLogRepository interface {
	CreateActivityLog(ctx context.Context, req *entity.ActivityLog) error
	GetActivityLog(ctx context.Context, req *entity.GetActivityLogReq) ([]entity.ActivityLog, error)
}

type ActivityLogService interface {
	CreateActivityLog(ctx context.Context, req *entity.ActivityLog) error
	GetActivityLog(ctx context.Context, req *entity.GetActivityLogReq) ([]entity.ActivityLog, error)
}
