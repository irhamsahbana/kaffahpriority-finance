package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/pkg/errmsg"
	"context"
	"database/sql"

	"github.com/rs/zerolog/log"
)

func (s *payrollService) GetPayrollItemDetail(ctx context.Context, req *entity.GetPayrollItemDetailReq) (*entity.GetPayrollItemDetailResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetPayrollItemDetail")
	defer span.End()

	item, err := s.repo.GetPayrollItem(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errmsg.NewCustomErrors(404).SetMessage("Data tidak ditemukan")
		}
		log.Ctx(ctx).Error().Err(err).Msg("failed to get payroll item")
		return nil, err
	}

	studentsMap, err := s.repo.GetAdditionalStudents(ctx, []string{item.ID})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get additional students")
		return nil, err
	}

	if students, ok := studentsMap[item.ID]; ok {
		item.AdditionalStudents = students
	} else {
		item.AdditionalStudents = make([]entity.PayrollItemAdditionalStudent, 0)
	}

	return &entity.GetPayrollItemDetailResp{
		PayrollItem: *item,
	}, nil
}
