package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/pkg"
	"context"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

func (s *payrollService) UpdatePayrollItem(ctx context.Context, req *entity.UpdatePayrollItemReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.UpdatePayrollItem")
	defer span.End()

	// Get item before update for logging and wage calculation
	itemBefore, err := s.repo.GetPayrollItem(ctx, req.ID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get payroll item")
		return err
	}

	// If fields that affect wage are updated, recalculate wage
	if req.ProgramMeetings != nil || req.IsMeetingFull != nil || req.ForeignLearningFee != nil || req.NightLearningFee != nil {
		// Apply updates to a temp item struct to calculate
		meetings := itemBefore.ProgramMeetings
		if req.ProgramMeetings != nil {
			meetings = *req.ProgramMeetings
		}

		isFull := itemBefore.IsMeetingFull
		if req.IsMeetingFull != nil {
			isFull = *req.IsMeetingFull
		}

		fl := itemBefore.ForeignLearningFee
		if req.ForeignLearningFee != nil {
			fl = *req.ForeignLearningFee
		}

		nl := itemBefore.NightLearningFee
		if req.NightLearningFee != nil {
			nl = *req.NightLearningFee
		}

		// Recalculate wage
		// WagePerMeeting and FullWage in DB already account for ITP multiplier

		var initialFee float64
		if isFull {
			initialFee = itemBefore.FullWage.InexactFloat64()
		} else {
			initialFee = itemBefore.WagePerMeeting.Mul(decimal.NewFromInt(int64(meetings))).InexactFloat64()
		}

		newWage := nl.Add(fl).Add(decimal.NewFromFloat(initialFee))
		if meetings < 1 {
			newWage = decimal.Zero
		}

		req.Wage = &newWage
	}

	if err := s.repo.UpdatePayrollItem(ctx, req); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to update payroll item")
		return err
	}

	// Get item after update for logging
	itemAfter, err := s.repo.GetPayrollItem(ctx, req.ID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get payroll item after update")
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
			Type:       entity.ActivityLogTypeUpdate,
			AuthorName: user.Name,
			AuthorRole: user.Role,
			CreatedAt:  time.Now(),
			CreatedBy:  req.UserID,
			Message:    "berhasil mengubah data item payroll",
			Before:     itemBefore,
			After:      itemAfter,
		})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).Msg("failed to create activity log")
		}
	}()

	return nil
}
