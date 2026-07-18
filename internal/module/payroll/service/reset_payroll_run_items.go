package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/pkg"
	"codebase-app/pkg/errmsg"
	"context"
	"net/http"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (s *payrollService) ResetPayrollRunItems(ctx context.Context, req *entity.ResetPayrollRunItemsReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.ResetPayrollRunItems")
	defer span.End()

	// 1. Get payroll run
	run, err := s.repo.GetPayrollRun(ctx, req.ID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get payroll run")
		return err
	}

	// 2. Validate draft status
	if run.Status != entity.PayrollRunStatusDraft {
		return errmsg.NewCustomErrors(http.StatusBadRequest,
			errmsg.WithMessage("Tidak dapat mereset data rekap gaji yang sudah diproses"),
		)
	}

	// 3. Validate current month
	timezone := "Asia/Makassar"
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to load timezone")
		return err
	}

	now := time.Now().In(loc)
	currentPeriod := now.Format("2006-01")

	if run.Period != currentPeriod {
		return errmsg.NewCustomErrors(http.StatusBadRequest,
			errmsg.WithMessage("Tidak dapat mereset data rekap gaji untuk bulan sebelumnya"),
		)
	}

	// 4. Reset payroll items
	rowsAffected, err := s.repo.ResetPayrollRunItems(ctx, req.ID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to reset payroll run items")
		return err
	}

	log.Ctx(ctx).Info().Int64("rows_affected", rowsAffected).Msg("payroll items reset successfully")

	// 5. Async Activity Log
	childCtx := pkg.GenerateChildContext(ctx, time.Minute)
	go func() {
		user, err := s.userRepo.GetMe(childCtx, &entity.GetMeReq{UserID: req.UserID})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).Msg("failed to get user data for activity log")
			return
		}

		err = s.activityLogRepo.CreateActivityLog(childCtx, &entity.ActivityLog{
			ID:         ulid.Make().String(),
			EntityName: entity.ActivityLogEntityPayrollRuns,
			EntityID:   req.ID,
			Type:       entity.ActivityLogTypeDelete,
			AuthorName: user.Name,
			AuthorRole: user.Role,
			CreatedAt:  time.Now(),
			CreatedBy:  req.UserID,
			Message:    "berhasil mereset data rekap gaji periode " + run.Period,
			Before:     run,
		})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).Msg("failed to create activity log")
		} else {
			log.Ctx(childCtx).Info().Msg("Activity log created successfully")
		}
	}()

	return nil
}
