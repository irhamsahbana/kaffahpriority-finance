package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/pkg"
	"time"

	portsActivityLog "codebase-app/internal/ports/module/activity_log"
	ports "codebase-app/internal/ports/module/report"
	portsUser "codebase-app/internal/ports/module/user"
	"context"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

var _ ports.ReportService = &reportService{}

type reportService struct {
	repo            ports.ReportRepository
	activityLogRepo portsActivityLog.ActivityLogRepository
	userRepo        portsUser.UserRepository
}

type Config struct {
	Repo            ports.ReportRepository
	ActivityLogRepo portsActivityLog.ActivityLogRepository
	UserRepo        portsUser.UserRepository
}

func NewReportService(cfg Config) *reportService {
	return &reportService{
		repo:            cfg.Repo,
		activityLogRepo: cfg.ActivityLogRepo,
		userRepo:        cfg.UserRepo,
	}
}

func (s *reportService) GetTemplates(ctx context.Context, req *entity.GetTemplatesReq) (*entity.GetTemplatesResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetTemplates")
	defer span.End()

	return s.repo.GetTemplates(ctx, req)
}

func (s *reportService) GetTemplate(ctx context.Context, req *entity.GetTemplateReq) (*entity.GetTemplateResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetTemplate")
	defer span.End()

	return s.repo.GetTemplate(ctx, req)
}

func (s *reportService) DeleteTemplate(ctx context.Context, req *entity.GetTemplateReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.DeleteTemplate")
	defer span.End()

	oldTemplate, err := s.repo.GetTemplate(ctx, req)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).
			Any(entity.Payload, req).
			Msgf("failed to get template")
		return err
	}

	err = s.repo.DeleteTemplate(ctx, req)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).
			Any(entity.Payload, req).
			Msgf("failed to delete template")
		return err
	}

	childCtx := pkg.GenerateChildContext(ctx, time.Minute/2)
	go func() {
		user, err := s.userRepo.GetMe(childCtx, &entity.GetMeReq{UserID: req.UserID})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get me")
			return
		}

		deleteTime := time.Now()
		err = s.activityLogRepo.CreateActivityLog(childCtx, &entity.ActivityLog{
			ID:         ulid.Make().String(),
			EntityName: entity.ActivityLogEntityProgramRegistrationTemplates,
			EntityID:   req.ID,
			Type:       entity.ActivityLogTypeDelete,
			AuthorName: user.Name,
			AuthorRole: user.Role,
			DeletedAt:  &deleteTime,
			DeletedBy:  &req.UserID,
			Message:    "berhasil menghapus template",
			Before:     oldTemplate,
		})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to create activity log")
		}
	}()

	return nil
}

func (s *reportService) CreateRegistrations(ctx context.Context, req *entity.CreateRegistrationsReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.CreateRegistrations")
	defer span.End()

	return s.repo.CreateRegistrations(ctx, req)
}

func (s *reportService) CopyRegistrations(ctx context.Context, req *entity.CopyRegistrationsReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.CopyRegistrations")
	defer span.End()

	return s.repo.CopyRegistrations(ctx, req)
}

func (s *reportService) UpdateRegistration(ctx context.Context, req *entity.UpdateRegistrationReq) (*entity.UpdateRegistrationResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.UpdateRegistration")
	defer span.End()

	// return s.repo.UpdateRegistration(ctx, req)
	oldRegistration, err := s.repo.GetRegistration(ctx, &entity.GetRegistrationReq{
		UserID: req.UserID,
		ID:     req.ID,
	})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).
			Any(entity.Payload, req).
			Msgf("failed to get registration")
		return nil, err
	}

	resp, err := s.repo.UpdateRegistrationV2(ctx, req)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).
			Any(entity.Payload, req).
			Msgf("failed to update registration")
		return nil, err
	}

	childCtx := pkg.GenerateChildContext(ctx, time.Minute/2)
	go func() {
		user, err := s.userRepo.GetMe(childCtx, &entity.GetMeReq{UserID: req.UserID})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get me")
			return
		}

		registration, err := s.repo.GetRegistration(childCtx, &entity.GetRegistrationReq{
			UserID: req.UserID,
			ID:     resp.ID,
		})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get registration")
			return
		}

		err = s.activityLogRepo.CreateActivityLog(childCtx, &entity.ActivityLog{
			ID:         ulid.Make().String(),
			EntityName: entity.ActivityLogEntityProgramRegistrations,
			EntityID:   req.ID,
			Type:       entity.ActivityLogTypeUpdate,
			AuthorName: user.Name,
			AuthorRole: user.Role,
			UpdatedAt:  time.Now(),
			UpdatedBy:  req.UserID,
			Message:    "berhasil mengupdate registrasi",
			Before:     oldRegistration,
			After:      registration,
		})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to create activity log")
		}
	}()

	return resp, nil
}

func (s *reportService) UpdateRegistrationLecturer(ctx context.Context, req *entity.UpdateRegistrationLecturerReq) (*entity.UpdateRegistrationLecturerResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.UpdateRegistrationLecturer")
	defer span.End()

	oldRegistration, err := s.repo.GetRegistration(ctx, &entity.GetRegistrationReq{
		UserID: req.UserID,
		ID:     req.ID,
	})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).
			Any(entity.Payload, req).
			Msgf("failed to get registration")
		return nil, err
	}

	resp, err := s.repo.UpdateRegistrationLecturer(ctx, req)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).
			Any(entity.Payload, req).
			Msgf("failed to update registration lecturer")
		return nil, err
	}

	childCtx := pkg.GenerateChildContext(ctx, time.Minute/2)
	go func() {
		user, err := s.userRepo.GetMe(childCtx, &entity.GetMeReq{UserID: req.UserID})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get me")
			return
		}

		registration, err := s.repo.GetRegistration(childCtx, &entity.GetRegistrationReq{
			UserID: req.UserID,
			ID:     req.ID,
		})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get registration")
			return
		}

		err = s.activityLogRepo.CreateActivityLog(childCtx, &entity.ActivityLog{
			ID:         ulid.Make().String(),
			EntityName: entity.ActivityLogEntityProgramRegistrations,
			EntityID:   req.ID,
			Type:       entity.ActivityLogTypeUpdate,
			AuthorName: user.Name,
			AuthorRole: user.Role,
			UpdatedAt:  time.Now(),
			UpdatedBy:  req.UserID,
			Message:    "berhasil mengupdate pengajar pada registrasi",
			Before:     oldRegistration,
			After:      registration,
		})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to create activity log")
		}
	}()

	return resp, nil
}

func (s *reportService) UpdateRegistrationIsPaid(ctx context.Context, req *entity.UpdateRegistrationIsPaidReq) (*entity.UpdateRegistrationIsPaidResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.UpdateRegistrationIsPaid")
	defer span.End()

	oldRegistration, err := s.repo.GetRegistration(ctx, &entity.GetRegistrationReq{
		UserID: req.UserID,
		ID:     req.ID,
	})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).
			Any(entity.Payload, req).
			Msgf("failed to get registration")
		return nil, err
	}

	resp, err := s.repo.UpdateRegistrationIsPaid(ctx, req)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).
			Any(entity.Payload, req).
			Msgf("failed to update registration is paid")
		return nil, err
	}

	childCtx := pkg.GenerateChildContext(ctx, time.Minute/2)
	go func() {
		user, err := s.userRepo.GetMe(childCtx, &entity.GetMeReq{UserID: req.UserID})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get me")
			return
		}

		registration, err := s.repo.GetRegistration(childCtx, &entity.GetRegistrationReq{
			UserID: req.UserID,
			ID:     req.ID,
		})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get registration")
			return
		}

		err = s.activityLogRepo.CreateActivityLog(childCtx, &entity.ActivityLog{
			ID:         ulid.Make().String(),
			EntityName: entity.ActivityLogEntityProgramRegistrations,
			EntityID:   req.ID,
			Type:       entity.ActivityLogTypeUpdate,
			AuthorName: user.Name,
			AuthorRole: user.Role,
			UpdatedAt:  time.Now(),
			UpdatedBy:  req.UserID,
			Message:    "berhasil mengupdate status pembayaran registrasi",
			Before:     oldRegistration,
			After:      registration,
		})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to create activity log")
		}
	}()

	return resp, nil
}

func (s *reportService) UpdateRegistrationsPaidAt(ctx context.Context, req *entity.UpdateRegisPaidAtReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.UpdateRegistrationsPaidAt")
	defer span.End()

	return s.repo.UpdateRegistrationsPaidAt(ctx, req)
}

func (s *reportService) GetRegistrations(ctx context.Context, req *entity.GetRegistrationsReq) (*entity.GetRegistrationsResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetRegistrations")
	defer span.End()

	return s.repo.GetRegistrations(ctx, req)
}

func (s *reportService) GetUnusedRegistrations(ctx context.Context, req *entity.GetRegistrationsReq) (*entity.GetRegistrationsResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetUnusedRegistrations")
	defer span.End()

	return s.repo.GetUnusedRegistrations(ctx, req)
}

func (s *reportService) GetRegistration(ctx context.Context, req *entity.GetRegistrationReq) (*entity.GetRegistrationResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetRegistration")
	defer span.End()

	return s.repo.GetRegistration(ctx, req)
}

func (s *reportService) DeleteRegistration(ctx context.Context, req *entity.GetRegistrationReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.DeleteRegistration")
	defer span.End()

	return s.repo.DeleteRegistration(ctx, req)
}

func (s *reportService) RegistrationsMarkAsUsed(ctx context.Context, req *entity.RegistrationsMarkAsUsedReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.RegistrationsMarkAsUsed")
	defer span.End()

	ids, err := s.repo.RegistrationsMarkAsUsed(ctx, req)
	if err != nil {
		return err
	}

	childCtx := pkg.GenerateChildContext(ctx, time.Minute/2)
	go func() {
		user, err := s.userRepo.GetMe(childCtx, &entity.GetMeReq{UserID: req.UserID})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get me")
			return
		}

		getRegReq := &entity.GetRegistrationsReq{
			UserID: req.UserID,
			IDs:    ids,
		}
		getRegReq.SetDefault()
		getRegReq.Paginate = len(ids)
		getRegReq.Page = 1

		registrations, err := s.repo.GetRegistrations(childCtx, getRegReq)
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get registrations")
			return
		}

		for _, registration := range registrations.Items {
			err = s.activityLogRepo.CreateActivityLog(childCtx, &entity.ActivityLog{
				ID:         ulid.Make().String(),
				EntityName: entity.ActivityLogEntityProgramRegistrations,
				EntityID:   registration.ID,
				Type:       entity.ActivityLogTypeUpdate,
				AuthorName: user.Name,
				AuthorRole: user.Role,
				UpdatedAt:  time.Now(),
				UpdatedBy:  req.UserID,
				Message:    "berhasil menandai registrasi sebagai terpakai",
				After:      registration,
			})
			if err != nil {
				log.Ctx(childCtx).Error().Err(err).
					Any(entity.Payload, req).
					Msgf("failed to create activity log")
			}
		}
	}()

	return nil
}

func (s *reportService) GetSummaries(ctx context.Context, req *entity.GetSummariesReq) (*entity.GetSummariesResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetSummaries")
	defer span.End()

	return s.repo.GetSummaries(ctx, req)
}

func (s *reportService) GetSummariesForCFO2(ctx context.Context, req *entity.GetSummariesReq) (*entity.GetSummariesForCFO2Resp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetSummariesForCFO2")
	defer span.End()

	return s.repo.GetSummariesForCFO2(ctx, req)
}

func (s *reportService) GetLecturerPrograms(ctx context.Context, req *entity.GetLecturerProgramsReq) (*entity.GetLecturerProgramsResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetLecturerPrograms")
	defer span.End()

	return s.repo.GetLecturerPrograms(ctx, req)
}

func (s *reportService) GetRegistrationsPerLecturer(ctx context.Context, req *entity.GetRegistrationListPerLecturerReq) (*entity.GetRegistrationListPerLecturerResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetRegistrationsPerLecturer")
	defer span.End()

	// return s.repo.GetRegistrationsPerLecturer(ctx, req)
	return s.repo.GetRegistrationsPerLecturerV2(ctx, req)
}

func (s *reportService) GetLecturersWages(ctx context.Context, req *entity.GetLecturersWagesReq) (*entity.GetLecturersWagesResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetLecturersWages")
	defer span.End()

	return s.repo.GetLecturersWages(ctx, req)
}

func (s *reportService) GetLecturersWagesAggregate(ctx context.Context, req *entity.GetLecturersWagesAggregateReq) (*entity.LecturersWageAggregateResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetLecturersWagesAggregate")
	defer span.End()

	return s.repo.GetLecturersWagesAggregate(ctx, req)
}

func (s *reportService) GetLecturersWagesAggregateYearly(ctx context.Context, req *entity.GetLecturersWagesAggregateYearlyReq) (*entity.LecturersWageAggregateYearlyResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetLecturersWagesAggregateYearly")
	defer span.End()

	return s.repo.GetLecturersWagesAggregateYearly(ctx, req)
}

func (s *reportService) UpdateLecturersWage(ctx context.Context, req *entity.UpdateLecturersWageReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.UpdateLecturersWage")
	defer span.End()

	oldRegistration, err := s.repo.GetRegistration(ctx, &entity.GetRegistrationReq{
		UserID: req.UserID,
		ID:     req.RegistrationID,
	})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).
			Any(entity.Payload, req).
			Msgf("failed to get registration")
		return err
	}

	err = s.repo.UpdateLecturersWage(ctx, req)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).
			Any(entity.Payload, req).
			Msgf("failed to update lecturers wage")
		return err
	}

	go func() {
		user, err := s.userRepo.GetMe(ctx, &entity.GetMeReq{UserID: req.UserID})
		if err != nil {
			log.Ctx(ctx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get me")
			return
		}

		registration, err := s.repo.GetRegistration(ctx, &entity.GetRegistrationReq{
			UserID: req.UserID,
			ID:     req.RegistrationID,
		})
		if err != nil {
			log.Ctx(ctx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get registration")
			return
		}

		err = s.activityLogRepo.CreateActivityLog(ctx, &entity.ActivityLog{
			ID:         ulid.Make().String(),
			EntityName: entity.ActivityLogEntityProgramRegistrations,
			EntityID:   req.RegistrationID,
			Type:       entity.ActivityLogTypeUpdate,
			AuthorName: user.Name,
			AuthorRole: user.Role,
			UpdatedAt:  time.Now(),
			UpdatedBy:  req.UserID,
			Message:    "berhasil mengupdate ujrah pengajar",
			Before:     oldRegistration,
			After:      registration,
		})
		if err != nil {
			log.Ctx(ctx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to create activity log")
		}
	}()

	return nil
}

func (s *reportService) BulkUpdateLecturersWage(ctx context.Context, req *entity.BulkUpdateLecturersWageReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.BulkUpdateLecturersWage")
	defer span.End()

	// TODO: implement bulk update activity log if needed
	return s.repo.BulkUpdateLecturersWage(ctx, req)
}

func (s *reportService) DistributeHRFee(ctx context.Context, req *entity.HRDistributionReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.DistributeHRFee")
	defer span.End()

	oldRegistration, err := s.repo.GetRegistration(ctx, &entity.GetRegistrationReq{
		UserID: req.UserID,
		ID:     req.RegistrationID,
	})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).
			Any(entity.Payload, req).
			Msgf("failed to get registration")
		return err
	}

	err = s.repo.DistributeHRFee(ctx, req)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).
			Any(entity.Payload, req).
			Msgf("failed to distribute hr fee")
		return err
	}

	childCtx := pkg.GenerateChildContext(ctx, time.Minute/2)
	go func() {
		user, err := s.userRepo.GetMe(childCtx, &entity.GetMeReq{UserID: req.UserID})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get me")
			return
		}

		registration, err := s.repo.GetRegistration(childCtx, &entity.GetRegistrationReq{
			UserID: req.UserID,
			ID:     req.RegistrationID,
		})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get registration")
			return
		}

		err = s.activityLogRepo.CreateActivityLog(childCtx, &entity.ActivityLog{
			ID:         ulid.Make().String(),
			EntityName: entity.ActivityLogEntityProgramRegistrations,
			EntityID:   req.RegistrationID,
			Type:       entity.ActivityLogTypeUpdate,
			AuthorName: user.Name,
			AuthorRole: user.Role,
			UpdatedAt:  time.Now(),
			UpdatedBy:  req.UserID,
			Message:    "berhasil mendistribusikan biaya SDM",
			Before:     oldRegistration,
			After:      registration,
		})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to create activity log")
		}
	}()

	return nil
}

func (s *reportService) UseHRfeeForLecturer(ctx context.Context, req *entity.UseHRfeeForLecturerReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.UseHRfeeForLecturer")
	defer span.End()

	oldRegistration, err := s.repo.GetRegistration(ctx, &entity.GetRegistrationReq{
		UserID: req.UserID,
		ID:     req.RegistrationID,
	})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).
			Any(entity.Payload, req).
			Msgf("failed to get registration")
		return err
	}

	err = s.repo.UseHRfeeForLecturer(ctx, req)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).
			Any(entity.Payload, req).
			Msgf("failed to use hr fee for lecturer")
		return err
	}

	childCtx := pkg.GenerateChildContext(ctx, time.Minute/2)
	go func() {
		user, err := s.userRepo.GetMe(childCtx, &entity.GetMeReq{UserID: req.UserID})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get me")
			return
		}

		registration, err := s.repo.GetRegistration(ctx, &entity.GetRegistrationReq{
			UserID: req.UserID,
			ID:     req.RegistrationID,
		})
		if err != nil {
			log.Ctx(ctx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get registration")
			return
		}

		err = s.activityLogRepo.CreateActivityLog(childCtx, &entity.ActivityLog{
			ID:         ulid.Make().String(),
			EntityName: entity.ActivityLogEntityProgramRegistrations,
			EntityID:   req.RegistrationID,
			Type:       entity.ActivityLogTypeUpdate,
			AuthorName: user.Name,
			AuthorRole: user.Role,
			UpdatedAt:  time.Now(),
			UpdatedBy:  req.UserID,
			Message:    "berhasil menggunakan biaya SDM untuk pengajar",
			Before:     oldRegistration,
			After:      registration,
		})
		if err != nil {
			log.Ctx(ctx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to create activity log")
		}
	}()

	return nil
}

func (s *reportService) BulkUseHRfeeForLecturer(ctx context.Context, req *entity.BulkUseHRfeeForLecturerReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.BulkUseHRfeeForLecturer")
	defer span.End()

	return s.repo.BulkUseHRfeeForLecturer(ctx, req)
}

func (s *reportService) GetRelatedRegistrations(ctx context.Context, req *entity.GetRelatedRegistrationsReq) (*entity.GetRelatedRegistrationsResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetRelatedRegistrations")
	defer span.End()

	return s.repo.GetRelatedRegistrations(ctx, req)
}

func (s *reportService) GetAcquisitionRightsAggregate(ctx context.Context, req *entity.GetAcquisitionRightsAggregateReq) (*entity.GetAcquisitionRightsAggregateResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetAcquisitionRightsAggregate")
	defer span.End()

	return s.repo.GetAcquisitionRightsAggregate(ctx, req)
}

func (s *reportService) GetAcquisitionRightsAggregateFromPayrollItems(ctx context.Context, req *entity.GetAcquisitionRightsAggregateReq) (*entity.GetAcquisitionRightsAggregateResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.GetAcquisitionRightsAggregateFromPayrollItems")
	defer span.End()

	return s.repo.GetAcquisitionRightsAggregateFromPayrollItems(ctx, req)
}

func (s *reportService) GenerateRegistrationReports(ctx context.Context, req *entity.GenerateRegistrationsReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.GenerateRegistrationReports")
	defer span.End()

	return s.repo.GenerateRegistrationReports(ctx, req)
}

func (s *reportService) RegistrationMultiAllocation(ctx context.Context, req *entity.RegistrationMuliAllocationReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.RegistrationMultiAllocation")
	defer span.End()

	template, err := s.repo.GetTemplate(ctx, &entity.GetTemplateReq{
		UserID: req.UserID,
		ID:     req.TemplateId,
	})
	if err != nil {
		return err
	}

	req.Template = template

	return s.repo.RegistrationMultiAllocation(ctx, req)
}

func (s *reportService) CreateAdditionalRegistration(ctx context.Context, req *entity.CreateAdditionalRegistrationReq) (*entity.CreateAdditionalRegistrationResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.CreateAdditionalRegistration")
	defer span.End()

	resp, err := s.repo.CreateAdditionalRegistration(ctx, req)
	if err != nil {
		return nil, err
	}

	childCtx := pkg.GenerateChildContext(ctx, time.Minute/2)
	go func() {
		user, err := s.userRepo.GetMe(childCtx, &entity.GetMeReq{UserID: req.UserID})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get me")
			return
		}

		registration, err := s.repo.GetRegistration(childCtx, &entity.GetRegistrationReq{
			UserID: req.UserID,
			ID:     resp.ID,
		})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get registration")
			return
		}

		err = s.activityLogRepo.CreateActivityLog(childCtx, &entity.ActivityLog{
			ID:         ulid.Make().String(),
			EntityName: entity.ActivityLogEntityProgramRegistrations,
			EntityID:   resp.ID,
			Type:       entity.ActivityLogTypeCreate,
			AuthorName: user.Name,
			AuthorRole: user.Role,
			CreatedAt:  time.Now(),
			CreatedBy:  req.UserID,
			Message:    "berhasil membuat registrasi tambahan",
			After:      registration,
		})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to create activity log")
		}
	}()

	return resp, nil
}
