package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/pkg"
	"codebase-app/pkg/errmsg"
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
	req.ForeignLearningFee = &fl

	nl := itemBefore.NightLearningFee
	if req.NightLearningFee != nil {
		nl = *req.NightLearningFee
	}
	req.NightLearningFee = &nl

	var initialWage decimal.Decimal
	if isFull {
		initialWage = itemBefore.FullWage
	} else {
		initialWage = itemBefore.WagePerMeeting.Mul(decimal.NewFromInt(int64(meetings)))
	}

	if meetings < 1 {
		initialWage = decimal.Zero
	}
	req.InitialWage = &initialWage

	if req.Wage == nil {
		req.Wage = &itemBefore.Wage
	}

	// if AcquisitionRights is not nil, and not equal to 0, then Wage must be greater than 0
	if req.AcquisitionRights != nil && *req.AcquisitionRights != uint64(0) && req.Wage.LessThanOrEqual(decimal.Zero) {
		return errmsg.NewCustomErrors(400).SetMessage("Keep Gaji harus terisi jika ingin mnegubah Hak")
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
