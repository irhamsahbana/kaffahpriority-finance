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

	resp, err := s.repo.GetPayrollRunDetail(ctx, req.ID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get payroll run detail")
		return nil, err
	}

	return resp, nil
}
