package handler

import (
	"codebase-app/internal/adapter"
	m "codebase-app/internal/middleware"
	"codebase-app/internal/module/report/entity"
	"codebase-app/internal/module/report/ports"
	"codebase-app/internal/module/report/repository"
	"codebase-app/internal/module/report/service"
	"codebase-app/pkg/errmsg"
	"codebase-app/pkg/response"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type reportHandler struct {
	service ports.ReportService
}

func NewReportHandler() *reportHandler {
	var (
		repo    = repository.NewReportRepository()
		svc     = service.NewReportService(repo)
		handler = new(reportHandler)
	)
	handler.service = svc

	return handler
}

func (h *reportHandler) Register(router fiber.Router) {
	router.Post("/templates", m.AuthBearer, h.createTemplate)
	router.Get("/templates", m.AuthBearer, h.getTemplates)
	router.Put("/templates/:id", m.AuthBearer, h.updateTemplate)
	router.Get("/templates/:id", m.AuthBearer, h.getTemplate)
	router.Delete("/templates/:id", m.AuthBearer, h.deleteTemplate)

	router.Post("/registrations", m.AuthBearer, h.createRegistrations)
	router.Post("/copy-registrations", m.AuthBearer, h.copyRegistrations)
	router.Get("/registration-summaries", m.AuthBearer, h.getSummaries)
	router.Get("/registration-summaries-for-cfo2", m.AuthBearer, h.getSummariesForCFO2)
	router.Get("/registrations", m.AuthBearer, h.getRegistrations)
	router.Get("/exported-registrations", m.AuthBearer, h.getExportedRegistrations)
	router.Get("/exported-registrations-for-cfo2-monthly", m.AuthBearer, h.getExportedRegistrationsForCFO2Monthly)
	router.Get("/exported-registrations-for-cfo2-yearly", m.AuthBearer, h.getExportedRegistrationsForCFO2Yearly)
	router.Get("/exported-registrations-for-wage-recap-monthly", m.AuthBearer, h.getExportedRegistrationsForWageRecapMonthly)

	router.Put("/registrations/:id", m.AuthBearer, h.updateRegistration)
	router.Get("/registrations/:id", m.AuthBearer, h.getRegistration)
	router.Delete("/registrations/:id", m.AuthBearer, h.deleteRegistration)
	router.Put("/registrations/:id/hr-fee-distributions", m.AuthBearer, h.hrDistributions)
	router.Put("/registrations/:id/lecturer-distributions", m.AuthBearer, h.lecturerDistributions)
	router.Put("/registrations/:id/lecturers", m.AuthBearer, h.updateRegistrationLecturer)
	router.Put("/registrations/:id/is-paid", m.AuthBearer, h.updateRegistrationIsPaid)

	router.Get("/registration-per-lecturers", m.AuthBearer, h.getRegistrationListPerLecturer)
	router.Patch("/lecturers-wages/:id", m.AuthBearer, h.updateLecturerWages)
	router.Get("/lecturers-wages", m.AuthBearer, h.getLecturerWages)
	router.Get("/lecturers-wages-aggregate", m.AuthBearer, h.getLecturerWagesAggregate)
	router.Get("/acquisition-rights-aggregate", m.AuthBearer, h.getAcquisitionRightsAggregate)

	router.Post("/generate-registration-reports", m.AuthBearer, h.generateRegistrationReports)
	router.Post("/multi-allocation-registrations", m.AuthBearer, h.registrationMultiAllocation)

	router.Get("/lecturer-programs", m.AuthBearer, h.getLecturerPrograms)
}

func (h *reportHandler) getSummaries(c *fiber.Ctx) error {
	var (
		req = new(entity.GetSummariesReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	req.UserId = l.GetUserId()

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::getSummaries - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getSummaries - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	if err := req.Validate(); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getSummaries - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetSummaries(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *reportHandler) getSummariesForCFO2(c *fiber.Ctx) error {
	var (
		req = new(entity.GetSummariesReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	req.UserId = l.GetUserId()

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::getSummariesForCFO2 - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getSummariesForCFO2 - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetSummariesForCFO2(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *reportHandler) getExportedRegistrations(c *fiber.Ctx) error {
	var (
		req = new(entity.GetExportedRegistrationsReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::getExportedRegistrations - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserId = l.GetUserId()
	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getExportedRegistrations - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetExportedRegistrations(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	defer func() {
		// delete file manually
		if err := os.Remove(resp.FilePath); err != nil {
			log.Warn().Err(err).Msg("handler::getExportedRegistrations - delete file error")
		}
	}()

	// download file
	if err := c.Download(resp.FilePath, resp.FileName); err != nil {
		log.Warn().Err(err).Msg("handler::getExportedRegistrations - download file error")
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error(err))
	}

	return nil
}

func (h *reportHandler) getExportedRegistrationsForCFO2Monthly(c *fiber.Ctx) error {
	var (
		req = new(entity.GetExportedRegistrationsForCFO2MonthlyReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::getExportedRegistrationsForCFO2Monthly - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserId = l.GetUserId()
	// req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getExportedRegistrationsForCFO2Monthly - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetExportedRegistrationsForCFO2Monthly(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	// delete file manually
	defer func() {
		if err := os.Remove(resp.FilePath); err != nil {
			log.Warn().Err(err).Msg("handler::getExportedRegistrationsForCFO2Monthly - delete file error")
		}
	}()

	// download file
	if err := c.Download(resp.FilePath, resp.FileName); err != nil {
		log.Warn().Err(err).Msg("handler::getExportedRegistrationsForCFO2Monthly - download file error")
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error(err))
	}

	return nil
}

func (h *reportHandler) getExportedRegistrationsForCFO2Yearly(c *fiber.Ctx) error {
	var (
		req = new(entity.GetExportedRegistrationsForCFO2YearlyReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::getExportedRegistrationsForCFO2Yearly - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserId = l.GetUserId()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getExportedRegistrationsForCFO2Yearly - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetExportedRegistrationsForCFO2Yearly(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	defer func() {
		if err := os.Remove(resp.FilePath); err != nil {
			log.Warn().Err(err).Msg("handler::getExportedRegistrationsForCFO2Yearly - delete file error")
		}
	}()

	if err := c.Download(resp.FilePath, resp.FileName); err != nil {
		log.Warn().Err(err).Msg("handler::getExportedRegistrationsForCFO2Yearly - download file error")
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error(err))
	}

	return nil
}

func (h *reportHandler) getExportedRegistrationsForWageRecapMonthly(c *fiber.Ctx) error {
	var (
		req = new(entity.GetExportedRegistrationsForWageRecapMonthlyReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::getExportedRegistrationsForX - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserId = l.GetUserId()
	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getExportedRegistrationsForX - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetExportedRegistrationsForWageRecapMonthly(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	defer func() {
		if err := os.Remove(resp.FilePath); err != nil {
			log.Warn().Err(err).Msg("handler::getExportedRegistrationsForX - delete file error")
		}
	}()

	if err := c.Download(resp.FilePath, resp.FileName); err != nil {
		log.Warn().Err(err).Msg("handler::getExportedRegistrationsForX - download file error")
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error(err))
	}

	return nil

}

func (h *reportHandler) getRegistration(c *fiber.Ctx) error {
	var (
		req = new(entity.GetRegistrationReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	req.UserId = l.GetUserId()
	req.Id = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getRegistration - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetRegistration(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *reportHandler) deleteRegistration(c *fiber.Ctx) error {
	var (
		req = new(entity.GetRegistrationReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	req.UserId = l.GetUserId()
	req.Id = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::deleteRegistration - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	err := h.service.DeleteRegistration(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(nil, ""))
}

func (h *reportHandler) getLecturerPrograms(c *fiber.Ctx) error {
	var (
		req = new(entity.GetLecturerProgramsReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::getLecturerPrograms - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.SetDefault()
	req.UserId = l.GetUserId()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getLecturerPrograms - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetLecturerPrograms(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *reportHandler) hrDistributions(c *fiber.Ctx) error {
	var (
		req = new(entity.HRDistributionReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::hrDistributions - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserId = l.GetUserId()
	req.RegistrationId = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::hrDistributions - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	err := h.service.DistributeHRFee(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(nil, ""))
}

func (h *reportHandler) lecturerDistributions(c *fiber.Ctx) error {
	var (
		req = new(entity.UseHRfeeForLecturerReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::lecturerDistributions - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserId = l.GetUserId()
	req.RegistrationId = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::lecturerDistributions - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	if err := req.Validate(); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::lecturerDistributions - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	err := h.service.UseHRfeeForLecturer(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(nil, ""))
}

func (h *reportHandler) getRegistrationListPerLecturer(c *fiber.Ctx) error {
	var (
		req = new(entity.GetRegistrationListPerLecturerReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::getRegistrationListPerLecturer - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserId = l.GetUserId()
	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getRegistrationListPerLecturer - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetRegistrationsPerLecturer(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *reportHandler) getLecturerWages(c *fiber.Ctx) error {
	var (
		req = new(entity.GetLecturersWagesReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::getLecturerWages - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserId = l.GetUserId()
	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getLecturerWages - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetLecturersWages(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *reportHandler) getLecturerWagesAggregate(c *fiber.Ctx) error {
	var (
		req = new(entity.GetLecturersWagesAggregateReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::getLecturerWagesAggregate - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserId = l.GetUserId()
	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getLecturerWagesAggregate - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetLecturersWagesAggregate(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *reportHandler) getAcquisitionRightsAggregate(c *fiber.Ctx) error {
	var (
		req = new(entity.GetAcquisitionRightsAggregateReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::getAcquisitionRightsAggregate - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserId = l.GetUserId()
	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getAcquisitionRightsAggregate - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetAcquisitionRightsAggregate(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *reportHandler) updateLecturerWages(c *fiber.Ctx) error {
	var (
		req = new(entity.UpdateLecturersWageReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::updateLecturerWages - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserId = l.GetUserId()
	req.RegistrationId = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::updateLecturerWages - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	if err := req.Validate(); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::updateLecturerWages - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	err := h.service.UpdateLecturersWage(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(nil, ""))
}

func (h *reportHandler) generateRegistrationReports(c *fiber.Ctx) error {
	var (
		req = new(entity.GenerateRegistrationsReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::generateRegistrationReports - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.SetDefault()
	req.UserId = l.GetUserId()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::generateRegistrationReports - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	err := h.service.GenerateRegistrationReports(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(nil, ""))
}

func (h *reportHandler) registrationMultiAllocation(c *fiber.Ctx) error {
	var (
		req = new(entity.RegistrationMuliAllocationReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::registrationMultiAllocation - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserId = l.GetUserId()
	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::registrationMultiAllocation - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	err := h.service.RegistrationMultiAllocation(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(nil, ""))
}
