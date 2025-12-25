package service

import (
	"codebase-app/internal/entity"
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

func (s *reportService) CreateTemplate(ctx context.Context, req *entity.CreateTemplateReq) (*entity.CreateTemplateResp, error) {
	resp, err := s.repo.CreateTemplate(ctx, req)
	if err != nil {
		return nil, err
	}

	go func() {
		user, err := s.userRepo.GetMe(ctx, &entity.GetMeReq{UserID: req.UserID})
		if err != nil {
			log.Error().Err(err).Msgf("reportService.CreateTemplate - failed to get me")
			return
		}

		template, err := s.repo.GetTemplate(ctx, &entity.GetTemplateReq{
			ID: resp.ID,
		})
		if err != nil {
			log.Error().Err(err).Msgf("reportService.CreateTemplate - failed to get template")
			return
		}

		err = s.activityLogRepo.CreateActivityLog(ctx, &entity.ActivityLog{
			ID:         ulid.Make().String(),
			EntityName: entity.ActivityLogEntityProgramRegistrationTemplates,
			EntityID:   resp.ID,
			Type:       entity.ActivityLogTypeCreate,
			AuthorName: user.Name,
			AuthorRole: user.Role,
			CreatedAt:  time.Now(),
			CreatedBy:  req.UserID,
			Message:    "berhasil membuat template",
			After:      template,
		})
		if err != nil {
			log.Error().Err(err).Msgf("reportService.CreateTemplate - failed to create activity log")
		}
	}()

	return resp, nil
}

func (s *reportService) UpdateTemplate(ctx context.Context, req *entity.UpdateTemplateGeneralReq) (*entity.UpdateTemplateResp, error) {
	oldTemplate, err := s.repo.GetTemplate(ctx, &entity.GetTemplateReq{
		ID: req.ID,
	})
	if err != nil {
		log.Error().Err(err).
			Any(entity.Payload, req).
			Msgf("failed to get template")
		return nil, err
	}

	resp, err := s.repo.UpdateTemplate(ctx, req)
	if err != nil {
		log.Error().Err(err).
			Any(entity.Payload, req).
			Msgf("failed to update template")
		return nil, err
	}

	go func() {
		user, err := s.userRepo.GetMe(ctx, &entity.GetMeReq{UserID: req.UserID})
		if err != nil {
			log.Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get me")
			return
		}

		template, err := s.repo.GetTemplate(ctx, &entity.GetTemplateReq{
			ID: resp.ID,
		})
		if err != nil {
			log.Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get template")
			return
		}
		err = s.activityLogRepo.CreateActivityLog(ctx, &entity.ActivityLog{
			ID:         ulid.Make().String(),
			EntityName: entity.ActivityLogEntityProgramRegistrationTemplates,
			EntityID:   resp.ID,
			Type:       entity.ActivityLogTypeUpdate,
			AuthorName: user.Name,
			AuthorRole: user.Role,
			UpdatedAt:  time.Now(),
			UpdatedBy:  req.UserID,
			Message:    "berhasil mengupdate template",
			Before:     oldTemplate,
			After:      template,
		})
		if err != nil {
			log.Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to create activity log")
		}
	}()

	return resp, nil
}

func (s *reportService) GetTemplates(ctx context.Context, req *entity.GetTemplatesReq) (*entity.GetTemplatesResp, error) {
	return s.repo.GetTemplates(ctx, req)
}

func (s *reportService) GetTemplate(ctx context.Context, req *entity.GetTemplateReq) (*entity.GetTemplateResp, error) {
	return s.repo.GetTemplate(ctx, req)
}

func (s *reportService) DeleteTemplate(ctx context.Context, req *entity.GetTemplateReq) error {
	oldTemplate, err := s.repo.GetTemplate(ctx, req)
	if err != nil {
		log.Error().Err(err).
			Any(entity.Payload, req).
			Msgf("failed to get template")
		return err
	}

	err = s.repo.DeleteTemplate(ctx, req)
	if err != nil {
		log.Error().Err(err).
			Any(entity.Payload, req).
			Msgf("failed to delete template")
		return err
	}

	go func() {
		user, err := s.userRepo.GetMe(ctx, &entity.GetMeReq{UserID: req.UserID})
		if err != nil {
			log.Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get me")
			return
		}

		deleteTime := time.Now()
		err = s.activityLogRepo.CreateActivityLog(ctx, &entity.ActivityLog{
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
			log.Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to create activity log")
		}
	}()

	return nil
}

func (s *reportService) CreateRegistrations(ctx context.Context, req *entity.CreateRegistrationsReq) error {
	return s.repo.CreateRegistrations(ctx, req)
}

func (s *reportService) CopyRegistrations(ctx context.Context, req *entity.CopyRegistrationsReq) error {
	return s.repo.CopyRegistrations(ctx, req)
}

func (s *reportService) UpdateRegistration(ctx context.Context, req *entity.UpdateRegistrationReq) (*entity.UpdateRegistrationResp, error) {
	// return s.repo.UpdateRegistration(ctx, req)
	oldRegistration, err := s.repo.GetRegistration(ctx, &entity.GetRegistrationReq{
		UserID: req.UserID,
		ID:     req.ID,
	})
	if err != nil {
		log.Error().Err(err).
			Any(entity.Payload, req).
			Msgf("failed to get registration")
		return nil, err
	}

	resp, err := s.repo.UpdateRegistration(ctx, req)
	if err != nil {
		log.Error().Err(err).
			Any(entity.Payload, req).
			Msgf("failed to update registration")
		return nil, err
	}

	go func() {
		user, err := s.userRepo.GetMe(ctx, &entity.GetMeReq{UserID: req.UserID})
		if err != nil {
			log.Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get me")
			return
		}

		registration, err := s.repo.GetRegistration(ctx, &entity.GetRegistrationReq{
			UserID: req.UserID,
			ID:     resp.ID,
		})
		if err != nil {
			log.Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get registration")
			return
		}

		err = s.activityLogRepo.CreateActivityLog(ctx, &entity.ActivityLog{
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
			log.Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to create activity log")
		}
	}()

	return resp, nil
}

func (s *reportService) UpdateRegistrationLecturer(ctx context.Context, req *entity.UpdateRegistrationLecturerReq) (*entity.UpdateRegistrationLecturerResp, error) {
	return s.repo.UpdateRegistrationLecturer(ctx, req)
}

func (s *reportService) UpdateRegistrationIsPaid(ctx context.Context, req *entity.UpdateRegistrationIsPaidReq) (*entity.UpdateRegistrationIsPaidResp, error) {
	return s.repo.UpdateRegistrationIsPaid(ctx, req)
}

func (s *reportService) UpdateRegistrationsPaidAt(ctx context.Context, req *entity.UpdateRegisPaidAtReq) error {
	return s.repo.UpdateRegistrationsPaidAt(ctx, req)
}

func (s *reportService) GetRegistrations(ctx context.Context, req *entity.GetRegistrationsReq) (*entity.GetRegistrationsResp, error) {
	return s.repo.GetRegistrations(ctx, req)
}

func (s *reportService) GetUnusedRegistrations(ctx context.Context, req *entity.GetRegistrationsReq) (*entity.GetRegistrationsResp, error) {
	return s.repo.GetUnusedRegistrations(ctx, req)
}

func (s *reportService) GetRegistration(ctx context.Context, req *entity.GetRegistrationReq) (*entity.GetRegistrationResp, error) {
	return s.repo.GetRegistration(ctx, req)
}

func (s *reportService) DeleteRegistration(ctx context.Context, req *entity.GetRegistrationReq) error {
	return s.repo.DeleteRegistration(ctx, req)
}

func (s *reportService) RegistrationsMarkAsUsed(ctx context.Context, req *entity.RegistrationsMarkAsUsedReq) error {
	return s.repo.RegistrationsMarkAsUsed(ctx, req)
}

func (s *reportService) GetSummaries(ctx context.Context, req *entity.GetSummariesReq) (*entity.GetSummariesResp, error) {
	return s.repo.GetSummaries(ctx, req)
}

func (s *reportService) GetSummariesForCFO2(ctx context.Context, req *entity.GetSummariesReq) (*entity.GetSummariesForCFO2Resp, error) {
	return s.repo.GetSummariesForCFO2(ctx, req)
}

func (s *reportService) GetLecturerPrograms(ctx context.Context, req *entity.GetLecturerProgramsReq) (*entity.GetLecturerProgramsResp, error) {
	return s.repo.GetLecturerPrograms(ctx, req)
}

func (s *reportService) GetRegistrationsPerLecturer(ctx context.Context, req *entity.GetRegistrationListPerLecturerReq) (*entity.GetRegistrationListPerLecturerResp, error) {
	// return s.repo.GetRegistrationsPerLecturer(ctx, req)
	return s.repo.GetRegistrationsPerLecturerV2(ctx, req)
}

func (s *reportService) GetLecturersWages(ctx context.Context, req *entity.GetLecturersWagesReq) (*entity.GetLecturersWagesResp, error) {
	return s.repo.GetLecturersWages(ctx, req)
}

func (s *reportService) GetLecturersWagesAggregate(ctx context.Context, req *entity.GetLecturersWagesAggregateReq) (*entity.LecturersWageAggregateResp, error) {
	return s.repo.GetLecturersWagesAggregate(ctx, req)
}

func (s *reportService) GetLecturersWagesAggregateYearly(ctx context.Context, req *entity.GetLecturersWagesAggregateYearlyReq) (*entity.LecturersWageAggregateYearlyResp, error) {
	return s.repo.GetLecturersWagesAggregateYearly(ctx, req)
}

func (s *reportService) UpdateLecturersWage(ctx context.Context, req *entity.UpdateLecturersWageReq) error {
	return s.repo.UpdateLecturersWage(ctx, req)
}

func (s *reportService) BulkUpdateLecturersWage(ctx context.Context, req *entity.BulkUpdateLecturersWageReq) error {
	return s.repo.BulkUpdateLecturersWage(ctx, req)
}

func (s *reportService) DistributeHRFee(ctx context.Context, req *entity.HRDistributionReq) error {
	return s.repo.DistributeHRFee(ctx, req)
}

func (s *reportService) UseHRfeeForLecturer(ctx context.Context, req *entity.UseHRfeeForLecturerReq) error {
	return s.repo.UseHRfeeForLecturer(ctx, req)
}

func (s *reportService) BulkUseHRfeeForLecturer(ctx context.Context, req *entity.BulkUseHRfeeForLecturerReq) error {
	return s.repo.BulkUseHRfeeForLecturer(ctx, req)
}

func (s *reportService) GetRelatedRegistrations(ctx context.Context, req *entity.GetRelatedRegistrationsReq) (*entity.GetRelatedRegistrationsResp, error) {
	return s.repo.GetRelatedRegistrations(ctx, req)
}

func (s *reportService) GetAcquisitionRightsAggregate(ctx context.Context, req *entity.GetAcquisitionRightsAggregateReq) (*entity.GetAcquisitionRightsAggregateResp, error) {
	return s.repo.GetAcquisitionRightsAggregate(ctx, req)
}

func (s *reportService) GenerateRegistrationReports(ctx context.Context, req *entity.GenerateRegistrationsReq) error {
	return s.repo.GenerateRegistrationReports(ctx, req)
}

func (s *reportService) RegistrationMultiAllocation(ctx context.Context, req *entity.RegistrationMuliAllocationReq) error {
	return s.repo.RegistrationMultiAllocation(ctx, req)
}

func (s *reportService) CreateAdditionalRegistration(ctx context.Context, req *entity.CreateAdditionalRegistrationReq) (*entity.CreateAdditionalRegistrationResp, error) {
	return s.repo.CreateAdditionalRegistration(ctx, req)
}
