package ports

import (
	"codebase-app/internal/module/report/entity"
	"context"
)

type ReportRepository interface {
	GetTemplates(ctx context.Context, req *entity.GetTemplatesReq) (*entity.GetTemplatesResp, error)
	GetTemplate(ctx context.Context, req *entity.GetTemplateReq) (*entity.GetTemplateResp, error)
	CreateTemplate(ctx context.Context, req *entity.CreateTemplateReq) (*entity.CreateTemplateResp, error)
	UpdateTemplate(ctx context.Context, req *entity.UpdateTemplateGeneralReq) (*entity.UpdateTemplateResp, error)
	DeleteTemplate(ctx context.Context, req *entity.GetTemplateReq) error

	CreateRegistrations(ctx context.Context, req *entity.CreateRegistrationsReq) error
	CopyRegistrations(ctx context.Context, req *entity.CopyRegistrationsReq) error
	GetRegistrations(ctx context.Context, req *entity.GetRegistrationsReq) (*entity.GetRegistrationsResp, error)
	GetUnusedRegistrations(ctx context.Context, req *entity.GetRegistrationsReq) (*entity.GetRegistrationsResp, error)
	GetRegistration(ctx context.Context, req *entity.GetRegistrationReq) (*entity.GetRegistrationResp, error)
	DeleteRegistration(ctx context.Context, req *entity.GetRegistrationReq) error
	UpdateRegistration(ctx context.Context, req *entity.UpdateRegistrationReq) (*entity.UpdateRegistrationResp, error)
	UpdateRegistrationLecturer(ctx context.Context, req *entity.UpdateRegistrationLecturerReq) (*entity.UpdateRegistrationLecturerResp, error)
	UpdateRegistrationIsPaid(ctx context.Context, req *entity.UpdateRegistrationIsPaidReq) (*entity.UpdateRegistrationIsPaidResp, error)
	UpdateRegistrationsPaidAt(ctx context.Context, req *entity.UpdateRegisPaidAtReq) error

	GetExportedRegistrations(ctx context.Context, req *entity.GetExportedRegistrationsReq) (*entity.GetExportedRegistrationsResp, error)
	GetExportedRegistrationsForCFO2Monthly(ctx context.Context, req *entity.GetExportedRegistrationsForCFO2MonthlyReq) (*entity.GetExportedRegistrationsForCFO2MonthlyResp, error)
	GetExportedRegistrationsForCFO2Yearly(ctx context.Context, req *entity.GetExportedRegistrationsForCFO2YearlyReq) (*entity.GetExportedRegistrationsForCFO2YearlyResp, error)
	GetExportedRegistrationsForWageRecapMonthly(ctx context.Context, req *entity.GetExportedRegistrationsForWageRecapMonthlyReq) (*entity.GetExportedRegistrationsForWageRecapMonthlyResp, error)
	GetExportedLecturersWages(ctx context.Context, req *entity.GetExportedLecturersWagesReq) (*entity.GetExportedLecturersWagesResp, error)

	ImportLecturersWages(ctx context.Context, req *entity.ImportLecturersWagesReq) error

	DistributeHRFee(ctx context.Context, req *entity.HRDistributionReq) error
	UseHRfeeForLecturer(ctx context.Context, req *entity.UseHRfeeForLecturerReq) error

	GetSummaries(ctx context.Context, req *entity.GetSummariesReq) (*entity.GetSummariesResp, error)
	GetSummariesForCFO2(ctx context.Context, req *entity.GetSummariesReq) (*entity.GetSummariesForCFO2Resp, error)
	GetLecturerPrograms(ctx context.Context, req *entity.GetLecturerProgramsReq) (*entity.GetLecturerProgramsResp, error)

	GetRegistrationsPerLecturer(ctx context.Context, req *entity.GetRegistrationListPerLecturerReq) (*entity.GetRegistrationListPerLecturerResp, error)
	GetLecturersWages(ctx context.Context, req *entity.GetLecturersWagesReq) (*entity.GetLecturersWagesResp, error)
	GetLecturersWagesAggregate(ctx context.Context, req *entity.GetLecturersWagesAggregateReq) (*entity.LecturersWageAggregateResp, error)
	UpdateLecturersWage(ctx context.Context, req *entity.UpdateLecturersWageReq) error

	GetAcquisitionRightsAggregate(ctx context.Context, req *entity.GetAcquisitionRightsAggregateReq) (*entity.GetAcquisitionRightsAggregateResp, error)

	GenerateRegistrationReports(ctx context.Context, req *entity.GenerateRegistrationsReq) error
	RegistrationMultiAllocation(ctx context.Context, req *entity.RegistrationMuliAllocationReq) error
}

type ReportService interface {
	GetTemplates(ctx context.Context, req *entity.GetTemplatesReq) (*entity.GetTemplatesResp, error)
	GetTemplate(ctx context.Context, req *entity.GetTemplateReq) (*entity.GetTemplateResp, error)
	CreateTemplate(ctx context.Context, req *entity.CreateTemplateReq) (*entity.CreateTemplateResp, error)
	UpdateTemplate(ctx context.Context, req *entity.UpdateTemplateGeneralReq) (*entity.UpdateTemplateResp, error)
	DeleteTemplate(ctx context.Context, req *entity.GetTemplateReq) error

	CreateRegistrations(ctx context.Context, req *entity.CreateRegistrationsReq) error
	CopyRegistrations(ctx context.Context, req *entity.CopyRegistrationsReq) error
	GetRegistrations(ctx context.Context, req *entity.GetRegistrationsReq) (*entity.GetRegistrationsResp, error)
	GetUnusedRegistrations(ctx context.Context, req *entity.GetRegistrationsReq) (*entity.GetRegistrationsResp, error)
	GetRegistration(ctx context.Context, req *entity.GetRegistrationReq) (*entity.GetRegistrationResp, error)
	DeleteRegistration(ctx context.Context, req *entity.GetRegistrationReq) error
	UpdateRegistration(ctx context.Context, req *entity.UpdateRegistrationReq) (*entity.UpdateRegistrationResp, error)
	UpdateRegistrationLecturer(ctx context.Context, req *entity.UpdateRegistrationLecturerReq) (*entity.UpdateRegistrationLecturerResp, error)
	UpdateRegistrationIsPaid(ctx context.Context, req *entity.UpdateRegistrationIsPaidReq) (*entity.UpdateRegistrationIsPaidResp, error)
	UpdateRegistrationsPaidAt(ctx context.Context, req *entity.UpdateRegisPaidAtReq) error

	GetExportedRegistrations(ctx context.Context, req *entity.GetExportedRegistrationsReq) (*entity.GetExportedRegistrationsResp, error)
	GetExportedRegistrationsForCFO2Monthly(ctx context.Context, req *entity.GetExportedRegistrationsForCFO2MonthlyReq) (*entity.GetExportedRegistrationsForCFO2MonthlyResp, error)
	GetExportedRegistrationsForCFO2Yearly(ctx context.Context, req *entity.GetExportedRegistrationsForCFO2YearlyReq) (*entity.GetExportedRegistrationsForCFO2YearlyResp, error)
	GetExportedRegistrationsForWageRecapMonthly(ctx context.Context, req *entity.GetExportedRegistrationsForWageRecapMonthlyReq) (*entity.GetExportedRegistrationsForWageRecapMonthlyResp, error)
	GetExportedLecturersWages(ctx context.Context, req *entity.GetExportedLecturersWagesReq) (*entity.GetExportedLecturersWagesResp, error)

	ImportLecturersWages(ctx context.Context, req *entity.ImportLecturersWagesReq) error

	DistributeHRFee(ctx context.Context, req *entity.HRDistributionReq) error
	UseHRfeeForLecturer(ctx context.Context, req *entity.UseHRfeeForLecturerReq) error

	GetSummaries(ctx context.Context, req *entity.GetSummariesReq) (*entity.GetSummariesResp, error)
	GetSummariesForCFO2(ctx context.Context, req *entity.GetSummariesReq) (*entity.GetSummariesForCFO2Resp, error)
	GetLecturerPrograms(ctx context.Context, req *entity.GetLecturerProgramsReq) (*entity.GetLecturerProgramsResp, error)

	GetRegistrationsPerLecturer(ctx context.Context, req *entity.GetRegistrationListPerLecturerReq) (*entity.GetRegistrationListPerLecturerResp, error)
	GetLecturersWages(ctx context.Context, req *entity.GetLecturersWagesReq) (*entity.GetLecturersWagesResp, error)
	GetLecturersWagesAggregate(ctx context.Context, req *entity.GetLecturersWagesAggregateReq) (*entity.LecturersWageAggregateResp, error)
	UpdateLecturersWage(ctx context.Context, req *entity.UpdateLecturersWageReq) error

	GetAcquisitionRightsAggregate(ctx context.Context, req *entity.GetAcquisitionRightsAggregateReq) (*entity.GetAcquisitionRightsAggregateResp, error)

	GenerateRegistrationReports(ctx context.Context, req *entity.GenerateRegistrationsReq) error
	RegistrationMultiAllocation(ctx context.Context, req *entity.RegistrationMuliAllocationReq) error
}
