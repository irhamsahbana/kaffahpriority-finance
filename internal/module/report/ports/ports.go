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

	CreateRegistrations(ctx context.Context, req *entity.CreateRegistrationsReq) error
	CopyRegistrations(ctx context.Context, req *entity.CopyRegistrationsReq) error
	GetRegistrations(ctx context.Context, req *entity.GetRegistrationsReq) (*entity.GetRegistrationsResp, error)
	GetRegistration(ctx context.Context, req *entity.GetRegistrationReq) (*entity.GetRegistrationResp, error)
	UpdateRegistration(ctx context.Context, req *entity.UpdateRegistrationReq) (*entity.UpdateRegistrationResp, error)

	DistributeHRFee(ctx context.Context, req *entity.HRDistributionReq) error
	UseHRfeeForLecturer(ctx context.Context, req *entity.UseHRfeeForLecturerReq) error

	GetSummaries(ctx context.Context, req *entity.GetSummariesReq) (*entity.GetSummariesResp, error)
	GetLecturerPrograms(ctx context.Context, req *entity.GetLecturerProgramsReq) (*entity.GetLecturerProgramsResp, error)

	GetRegistrationsPerLecturer(ctx context.Context, req *entity.GetRegistrationListPerLecturerReq) (*entity.GetRegistrationListPerLecturerResp, error)
}

type ReportService interface {
	GetTemplates(ctx context.Context, req *entity.GetTemplatesReq) (*entity.GetTemplatesResp, error)
	GetTemplate(ctx context.Context, req *entity.GetTemplateReq) (*entity.GetTemplateResp, error)
	CreateTemplate(ctx context.Context, req *entity.CreateTemplateReq) (*entity.CreateTemplateResp, error)
	UpdateTemplate(ctx context.Context, req *entity.UpdateTemplateGeneralReq) (*entity.UpdateTemplateResp, error)

	CreateRegistrations(ctx context.Context, req *entity.CreateRegistrationsReq) error
	CopyRegistrations(ctx context.Context, req *entity.CopyRegistrationsReq) error
	GetRegistrations(ctx context.Context, req *entity.GetRegistrationsReq) (*entity.GetRegistrationsResp, error)
	GetRegistration(ctx context.Context, req *entity.GetRegistrationReq) (*entity.GetRegistrationResp, error)
	UpdateRegistration(ctx context.Context, req *entity.UpdateRegistrationReq) (*entity.UpdateRegistrationResp, error)

	DistributeHRFee(ctx context.Context, req *entity.HRDistributionReq) error
	UseHRfeeForLecturer(ctx context.Context, req *entity.UseHRfeeForLecturerReq) error

	GetSummaries(ctx context.Context, req *entity.GetSummariesReq) (*entity.GetSummariesResp, error)
	GetLecturerPrograms(ctx context.Context, req *entity.GetLecturerProgramsReq) (*entity.GetLecturerProgramsResp, error)

	GetRegistrationsPerLecturer(ctx context.Context, req *entity.GetRegistrationListPerLecturerReq) (*entity.GetRegistrationListPerLecturerResp, error)
}
