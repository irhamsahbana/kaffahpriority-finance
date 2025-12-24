package service

import (
	"codebase-app/internal/entity"
	ports "codebase-app/internal/ports/module/activity_log"
	"context"
)

var _ ports.ActivityLogService = &activityLogService{}

type activityLogService struct {
	repo ports.ActivityLogRepository
}

func NewActivityLogService(repo ports.ActivityLogRepository) *activityLogService {
	return &activityLogService{
		repo: repo,
	}
}

func (s *activityLogService) CreateActivityLog(ctx context.Context, req *entity.ActivityLog) error {
	return s.repo.CreateActivityLog(ctx, req)
}

func (s *activityLogService) GetActivityLog(ctx context.Context, req *entity.GetActivityLogReq) ([]entity.ActivityLog, error) {
	return s.repo.GetActivityLog(ctx, req)
}
