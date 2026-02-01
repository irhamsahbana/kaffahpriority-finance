package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/pkg"
	"codebase-app/pkg/errmsg"
	"context"
	"fmt"
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

	if req.LecturerID != nil {
		lecturer, err := s.masterRepo.GetLecturer(ctx, &entity.GetLecturerReq{ID: *req.LecturerID})
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to get lecturer")
			return err
		}
		req.LecturerName = &lecturer.Name
	}

	if req.ProgramID != nil {
		program, err := s.masterRepo.GetProgram(ctx, &entity.GetProgramReq{ID: *req.ProgramID})
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to get program")
			return err
		}
		req.ProgramName = &program.Name
	}

	meetings := itemBefore.ProgramMeetings
	if req.ProgramMeetings != nil {
		meetings = *req.ProgramMeetings
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

	if req.InitialWage != nil {
		if meetings < 1 {
			zero := decimal.Zero
			req.InitialWage = &zero
		}
	}

	if req.Wage == nil {
		req.Wage = &itemBefore.Wage
	}

	if err := validatePayrollAdditionalStudents(req.AdditionalStudents); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("invalid additional students for payroll item")
		return err
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

func validatePayrollAdditionalStudents(students []entity.AddStudent) error {
	if students == nil {
		return nil
	}

	errs := errmsg.NewCustomErrors(400)
	for i, s := range students {
		if s.StudentID != nil && s.Name != nil {
			_ = errs.Add(fmt.Sprintf("additional_students[%d].student_id", i), "student_id dan name tidak boleh diisi bersamaan")
			_ = errs.Add(fmt.Sprintf("additional_students[%d].name", i), "student_id dan name tidak boleh diisi bersamaan")
		}
	}

	if errs.HasErrors() {
		return errs
	}

	return nil
}
