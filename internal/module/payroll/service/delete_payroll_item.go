package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/pkg"
	"context"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (s *payrollService) DeletePayrollItem(ctx context.Context, req *entity.DeletePayrollItemReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.DeletePayrollItem")
	defer span.End()

	// Get item before delete for logging
	itemBefore, err := s.repo.GetPayrollItem(ctx, req.ID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get payroll item")
		return err
	}

	if err := s.repo.DeletePayrollItem(ctx, req.ID); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to delete payroll item")
		return err
	}

	// Async Activity Log
	childCtx := pkg.GenerateChildContext(ctx, time.Minute)
	go func() {
		user, err := s.userRepo.GetMe(childCtx, &entity.GetMeReq{UserID: req.UserID})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).Msg("failed to get user data for activity log")
			return
		}

		err = s.activityLogRepo.CreateActivityLog(childCtx, &entity.ActivityLog{
			ID:         ulid.Make().String(),
			EntityName: entity.ActivityLogEntityPayrollItems,
			EntityID:   req.ID,
			Type:       entity.ActivityLogTypeDelete,
			AuthorName: user.Name,
			AuthorRole: user.Role,
			CreatedAt:  time.Now(),
			CreatedBy:  req.UserID,
			Message:    "berhasil menghapus item payroll",
			Before:     itemBefore,
		})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).Msg("failed to create activity log")
		}
	}()

	return nil
}
