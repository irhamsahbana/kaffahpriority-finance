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

func (s *payrollService) CreatePayrollItem(ctx context.Context, req *entity.CreatePayrollItemReq) (*entity.CreatePayrollItemResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.CreatePayrollItem")
	defer span.End()

	resp, err := s.repo.CreatePayrollItem(ctx, req)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to create payroll item")
		return nil, err
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
			EntityName: entity.ActivityLogEntityPayrollItems, // assuming payroll_items
			EntityID:   resp.ID,
			Type:       entity.ActivityLogTypeCreate,
			AuthorName: user.Name,
			AuthorRole: user.Role,
			CreatedAt:  time.Now(),
			CreatedBy:  req.UserID,
			Message:    "berhasil menambahkan payroll item untuk periode " + req.Period,
			After:      resp,
		})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).Msg("failed to create activity log")
		} else {
			log.Ctx(childCtx).Info().Msg("Activity log created successfully")
		}
	}()

	return resp, nil
}
