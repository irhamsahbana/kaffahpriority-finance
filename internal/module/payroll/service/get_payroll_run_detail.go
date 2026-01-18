package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"

	"github.com/rs/zerolog/log"
)

func (s *payrollService) GetPayrollRunDetail(ctx context.Context, req *entity.GetPayrollRunDetailReq) (*entity.GetPayrollRunDetailResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetPayrollRunDetail")
	defer span.End()

	run, err := s.repo.GetPayrollRun(ctx, req.ID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get payroll run")
		return nil, err
	}

	items, err := s.repo.GetPayrollItems(ctx, req.ID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get payroll items")
		return nil, err
	}

	return &entity.GetPayrollRunDetailResp{
		PayrollRun: *run,
		Items:      items,
	}, nil
}
