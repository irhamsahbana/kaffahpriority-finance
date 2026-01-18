package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"

	"github.com/rs/zerolog/log"
)

func (s *payrollService) GetPayrollRuns(ctx context.Context, req *entity.GetPayrollRunsReq) (*entity.GetPayrollRunsResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetPayrollRuns")
	defer span.End()

	items, total, err := s.repo.GetPayrollRuns(ctx, req)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get payroll runs")
		return nil, err
	}

	resp := &entity.GetPayrollRunsResp{
		Items: items,
	}
	resp.Meta.CountTotalPage(req.Page, req.Paginate, total)

	return resp, nil
}
