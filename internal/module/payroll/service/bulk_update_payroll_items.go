package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

func (s *payrollService) BulkUpdatePayrollItems(ctx context.Context, req *entity.BulkUpdatePayrollItemReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.BulkUpdatePayrollItems")
	defer span.End()

	// Pre-process items to calculate wages
	for i, itemReq := range req.Data {
		// Only fetch and recalculate if wage-affecting fields are present
		if itemReq.ProgramMeetings != nil || itemReq.IsMeetingFull != nil || itemReq.ForeignLearningFee != nil || itemReq.NightLearningFee != nil {

			itemBefore, err := s.repo.GetPayrollItem(ctx, itemReq.ID)
			if err != nil {
				log.Ctx(ctx).Error().Err(err).Str("id", itemReq.ID).Msg("failed to get payroll item for bulk update")
				return err
			}

			// Apply updates to temp variables
			meetings := itemBefore.ProgramMeetings
			if itemReq.ProgramMeetings != nil {
				meetings = *itemReq.ProgramMeetings
			}

			isFull := itemBefore.IsMeetingFull
			if itemReq.IsMeetingFull != nil {
				isFull = *itemReq.IsMeetingFull
			}

			fl := itemBefore.ForeignLearningFee
			if itemReq.ForeignLearningFee != nil {
				fl = *itemReq.ForeignLearningFee
			}

			nl := itemBefore.NightLearningFee
			if itemReq.NightLearningFee != nil {
				nl = *itemReq.NightLearningFee
			}

			// Recalculate wage
			var initialFee float64
			if isFull {
				initialFee = itemBefore.FullWage.InexactFloat64()
			} else {
				initialFee = itemBefore.WagePerMeeting.Mul(decimal.NewFromInt(int64(meetings))).InexactFloat64()
			}

			initialWageVal := decimal.NewFromFloat(initialFee)
			if meetings < 1 {
				initialWageVal = decimal.Zero
			}
			// Update the request object in the slice
			req.Data[i].InitialWage = &initialWageVal

			newWage := nl.Add(fl).Add(decimal.NewFromFloat(initialFee))
			if meetings < 1 {
				newWage = decimal.Zero
			}

			req.Data[i].Wage = &newWage
		}
	}

	return s.repo.BulkUpdatePayrollItems(ctx, req)
}
