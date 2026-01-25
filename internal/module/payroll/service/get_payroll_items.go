package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/pkg/types"
	"context"
	"database/sql"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

func (s *payrollService) GetPayrollItems(ctx context.Context, req *entity.GetPayrollItemsReq) (*entity.GetPayrollItemsResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetPayrollItems")
	defer span.End()

	var run *entity.PayrollRun
	var err error

	if req.Period != "" {
		run, err = s.repo.GetPayrollRunByPeriod(ctx, req.Period)
	} else {
		run, err = s.repo.GetLatestPayrollRun(ctx)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return &entity.GetPayrollItemsResp{
				Items: []entity.PayrollItem{},
				Meta: types.Meta{
					Page:      req.Page,
					Paginate:  req.Paginate,
					TotalData: 0,
					TotalPage: 1,
				},
			}, nil
		}
		log.Ctx(ctx).Error().Err(err).Msg("failed to get payroll run")
		return nil, err
	}

	items, total, err := s.repo.GetPayrollItemsWithPagination(ctx, run.ID, req)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get payroll items")
		return nil, err
	}

	// Enrich items with additional students and format names
	if len(items) > 0 {
		var itemIDs []string
		for _, item := range items {
			itemIDs = append(itemIDs, item.ID)
		}

		additionalStudentsMap, err := s.repo.GetAdditionalStudents(ctx, itemIDs)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to get additional students")
			return nil, err
		}

		for i := range items {
			// Append Additional Students
			if students, ok := additionalStudentsMap[items[i].ID]; ok {
				for _, s := range students {
					items[i].StudentName += ", " + s.Name
				}
			}

			// Append Program Flags
			if items[i].ForeignLearningFee.GreaterThan(decimal.Zero) {
				items[i].ProgramName += " + FL"
			}
			if items[i].NightLearningFee.GreaterThan(decimal.Zero) {
				items[i].ProgramName += " + NL"
			}
			if items[i].IsITP {
				items[i].ProgramName += " + ITP"
			}
			if items[i].Wage.LessThanOrEqual(decimal.Zero) {
				items[i].AcquisitionRights = 0
			}
		}
	}

	resp := &entity.GetPayrollItemsResp{
		Items: items,
	}
	resp.Meta.CountTotalPage(req.Page, req.Paginate, total)

	return resp, nil
}
