package service

import (
	"codebase-app/internal/module/report/entity"
	"codebase-app/internal/module/report/ports"
	"context"
)

var _ ports.ReportService = &reportService{}

type reportService struct {
	repo ports.ReportRepository
}

func NewReportService(repo ports.ReportRepository) *reportService {
	return &reportService{
		repo: repo,
	}
}

func (s *reportService) CreateTemplate(ctx context.Context, req *entity.CreateTemplateReq) (*entity.CreateTemplateResp, error) {
	return s.repo.CreateTemplate(ctx, req)
}

func (s *reportService) UpdateTemplate(ctx context.Context, req *entity.UpdateTemplateGeneralReq) (*entity.UpdateTemplateResp, error) {
	return s.repo.UpdateTemplate(ctx, req)
}

func (s *reportService) GetTemplates(ctx context.Context, req *entity.GetTemplatesReq) (*entity.GetTemplatesResp, error) {
	return s.repo.GetTemplates(ctx, req)
}

func (s *reportService) GetTemplate(ctx context.Context, req *entity.GetTemplateReq) (*entity.GetTemplateResp, error) {
	return s.repo.GetTemplate(ctx, req)
}

func (s *reportService) DeleteTemplate(ctx context.Context, req *entity.GetTemplateReq) error {
	return s.repo.DeleteTemplate(ctx, req)
}

func (s *reportService) CreateRegistrations(ctx context.Context, req *entity.CreateRegistrationsReq) error {
	return s.repo.CreateRegistrations(ctx, req)
}

func (s *reportService) CopyRegistrations(ctx context.Context, req *entity.CopyRegistrationsReq) error {
	return s.repo.CopyRegistrations(ctx, req)
}

func (s *reportService) UpdateRegistration(ctx context.Context, req *entity.UpdateRegistrationReq) (*entity.UpdateRegistrationResp, error) {
	return s.repo.UpdateRegistration(ctx, req)
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
