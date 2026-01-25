package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
)

func (s *payrollService) GetPayrollItemsPeriodically(ctx context.Context, req *entity.GetPayrollItemsPeriodicallyReq) (*entity.GetPayrollItemsPeriodicallyResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetPayrollItemsPeriodically")
	defer span.End()

	return s.repo.GetPayrollItemsPeriodically(ctx, req)
}
