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

func (s *payrollService) CreatePayrollRun(ctx context.Context, req *entity.CreatePayrollRunReq) (*entity.CreatePayrollRunResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.CreatePayrollRun")
	defer span.End()

	// 1. Determine period
	timezone := "Asia/Makassar"
	loc, _ := time.LoadLocation(timezone)
	now := time.Now().In(loc)
	period := now.Format("2006-01")

	// If period is provided in request, use it (optional feature, but good for testing)
	if req.Period != "" {
		period = req.Period
	}

	// 2. Generate Payroll Run
	run, err := s.repo.GeneratePayrollRun(ctx, period, timezone)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to generate payroll run")
		return nil, err
	}

	// Async Activity Log
	// Create a detached context for async logging
	childCtx := pkg.GenerateChildContext(ctx, time.Minute)
	go func() {
		log.Ctx(childCtx).Info().Msgf("Starting async activity log for payroll run: %s, user: %s", run.ID, req.UserID)

		user, err := s.userRepo.GetMe(childCtx, &entity.GetMeReq{UserID: req.UserID})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).Msg("failed to get user data for activity log")
			return
		}

		log.Ctx(childCtx).Info().Msgf("User found for activity log: %s", user.Name)

		err = s.activityLogRepo.CreateActivityLog(childCtx, &entity.ActivityLog{
			ID:         ulid.Make().String(),
			EntityName: entity.ActivityLogEntityPayrollRuns,
			EntityID:   run.ID,
			Type:       entity.ActivityLogTypeUpsert,
			AuthorName: user.Name,
			AuthorRole: user.Role,
			CreatedAt:  time.Now(),
			CreatedBy:  req.UserID,
			Message:    "berhasil membuat payroll run periode " + period,
			After:      run,
		})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).Msg("failed to create activity log")
		} else {
			log.Ctx(childCtx).Info().Msg("Activity log created successfully")
		}
	}()

	return &entity.CreatePayrollRunResp{
		ID: run.ID,
	}, nil
}
