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

	for i, itemReq := range req.Data {
		if err := validatePayrollAdditionalStudents(itemReq.AdditionalStudents); err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("id", itemReq.ID).Msg("invalid additional students for payroll item")
			return err
		}

		itemBefore, err := s.repo.GetPayrollItem(ctx, itemReq.ID)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Str("id", itemReq.ID).Msg("failed to get payroll item for bulk update")
			return err
		}

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
		req.Data[i].ForeignLearningFee = &fl

		nl := itemBefore.NightLearningFee
		if itemReq.NightLearningFee != nil {
			nl = *itemReq.NightLearningFee
		}
		req.Data[i].NightLearningFee = &nl

		var initialWage decimal.Decimal
		if isFull {
			initialWage = itemBefore.FullWage
		} else {
			initialWage = itemBefore.WagePerMeeting.Mul(decimal.NewFromInt(int64(meetings)))
		}

		if meetings < 1 {
			initialWage = decimal.Zero
		}
		req.Data[i].InitialWage = &initialWage

		if req.Data[i].Wage == nil {
			req.Data[i].Wage = &itemBefore.Wage
		}
	}

	return s.repo.BulkUpdatePayrollItems(ctx, req)
}
