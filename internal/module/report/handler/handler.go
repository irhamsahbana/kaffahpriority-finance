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
	protected := router.Group("/", m.AuthBearer)

	protected.Post("/templates", h.createTemplate)
	protected.Get("/templates", h.getTemplates)
	protected.Put("/templates/:id", h.updateTemplate)
	protected.Get("/templates/:id", h.getTemplate)
	protected.Delete("/templates/:id", h.deleteTemplate)

	// protected.Post("/registrations", h.createRegistrations)
	protected.Post("/copy-registrations", h.copyRegistrations)
	protected.Get("/registration-summaries", h.getSummaries)
	protected.Get("/registration-summaries-for-cfo2", h.getSummariesForCFO2)
	protected.Patch("/registration-paid-at-attributes", h.updateRegistrationsPaidAt)

	protected.Get("/registrations", h.getRegistrations)
	protected.Get("/unused-registrations", h.getUnusedRegistrations)
	protected.Get("/exported-registrations", h.getExportedRegistrations)
	protected.Get("/exported-registrations-for-cfo2-monthly", h.getExportedRegistrationsForCFO2Monthly)
	protected.Get("/exported-registrations-for-cfo2-yearly", h.getExportedRegistrationsForCFO2Yearly)
	protected.Get("/exported-registrations-for-wage-recap-monthly", h.getExportedRegistrationsForWageRecapMonthly)
	protected.Get("/exported-lecturers-wages", h.getExportedLecturersWages)

	protected.Post("import-lecturers-wages", h.importLecturersWages)

	protected.Put("/registrations/:id", h.updateRegistration)
	protected.Get("/registrations/:id", h.getRegistration)
	protected.Delete("/registrations/:id", h.deleteRegistration)
	protected.Put("/registrations/:id/hr-fee-distributions", h.hrDistributions)
	protected.Put("/registrations/:id/lecturer-distributions", h.lecturerDistributions)
	protected.Get("/registrations/:id/related-registrations", h.getRelatedRegistrations)
	protected.Put("/registrations/:id/lecturers", h.updateRegistrationLecturer)
	protected.Put("/registrations/:id/is-paid", h.updateRegistrationIsPaid)

	protected.Post("/registrations/:id/additional", h.createAdditionalRegistration)

	protected.Get("/registration-per-lecturers", h.getRegistrationListPerLecturer)
	protected.Patch("/lecturers-wages/:id", h.updateLecturerWages)
	protected.Get("/lecturers-wages", h.getLecturerWages)
	protected.Get("/lecturers-wages-aggregate", h.getLecturerWagesAggregate)
	protected.Get("/lecturers-wages-aggregate-yearly", h.getLecturerWagesAggregateYearly)
	protected.Get("/acquisition-rights-aggregate", h.getAcquisitionRightsAggregate)

	protected.Post("/generate-registration-reports", h.generateRegistrationReports)
	protected.Post("/multi-allocation-registrations", h.registrationMultiAllocation)

	protected.Get("/lecturer-programs", h.getLecturerPrograms)
}

func (h *reportHandler) getSummaries(c *fiber.Ctx) error {
	var (
		fnName = "handler::getSummaries"
		req    = new(entity.GetSummariesReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	req.UserID = l.GetUserId()

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	if err := req.Validate(); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
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
		fnName = "handler::getSummariesForCFO2"
		req    = new(entity.GetSummariesReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	req.UserID = l.GetUserId()

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
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

func (h *reportHandler) updateRegistrationsPaidAt(c *fiber.Ctx) error {
	var (
		fnName = "handler::updateRegistrationsPaidAt"
		req    = new(entity.UpdateRegisPaidAtReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	err := h.service.UpdateRegistrationsPaidAt(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(nil, ""))
}

func (h *reportHandler) getExportedRegistrations(c *fiber.Ctx) error {
	var (
		fnName = "handler::getExportedRegistrations"
		req    = new(entity.GetExportedRegistrationsReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()
	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
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
			log.Warn().Err(err).Msgf("%s - delete file error", fnName)
		}
	}()

	// download file
	if err := c.Download(resp.FilePath, resp.FileName); err != nil {
		log.Warn().Err(err).Msgf("%s - download file error", fnName)
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error(err))
	}

	return nil
}

func (h *reportHandler) createAdditionalRegistration(c *fiber.Ctx) error {
	var (
		fnName = "handler::createAdditionalRegistration"
		req    = new(entity.CreateAdditionalRegistrationReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()
	req.ParentID = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.CreateAdditionalRegistration(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success(resp, ""))
}

func (h *reportHandler) getExportedRegistrationsForCFO2Monthly(c *fiber.Ctx) error {
	var (
		fnName = "handler::getExportedRegistrationsForCFO2Monthly"
		req    = new(entity.GetExportedRegistrationsForCFO2MonthlyReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()
	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
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
			log.Warn().Err(err).Msgf("%s - delete file error", fnName)
		}
	}()

	// download file
	if err := c.Download(resp.FilePath, resp.FileName); err != nil {
		log.Warn().Err(err).Msgf("%s - download file error", fnName)
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error(err))
	}

	return nil
}

func (h *reportHandler) getExportedRegistrationsForCFO2Yearly(c *fiber.Ctx) error {
	var (
		fnName = "handler::getExportedRegistrationsForCFO2Yearly"
		req    = new(entity.GetExportedRegistrationsForCFO2YearlyReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
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
			log.Warn().Err(err).Msgf("%s - delete file error", fnName)
		}
	}()

	if err := c.Download(resp.FilePath, resp.FileName); err != nil {
		log.Warn().Err(err).Msgf("%s - download file error", fnName)
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error(err))
	}

	return nil
}

func (h *reportHandler) getExportedRegistrationsForWageRecapMonthly(c *fiber.Ctx) error {
	var (
		fnName = "handler::getExportedRegistrationsForWageRecapMonthly"
		req    = new(entity.GetExportedRegistrationsForWageRecapMonthlyReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()
	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
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
			log.Warn().Err(err).Msgf("%s - delete file error", fnName)
		}
	}()

	if err := c.Download(resp.FilePath, resp.FileName); err != nil {
		log.Warn().Err(err).Msgf("%s - download file error", fnName)
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error(err))
	}

	return nil

}

func (h *reportHandler) getRegistration(c *fiber.Ctx) error {
	var (
		fnName = "handler::getRegistration"
		req    = new(entity.GetRegistrationReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	req.UserID = l.GetUserId()
	req.ID = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
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
		fnName = "handler::deleteRegistration"
		req    = new(entity.GetRegistrationReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	req.UserID = l.GetUserId()
	req.ID = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
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
		fnName = "handler::getLecturerPrograms"
		req    = new(entity.GetLecturerProgramsReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.SetDefault()
	req.UserID = l.GetUserId()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
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
		fnName = "handler::hrDistributions"
		req    = new(entity.HRDistributionReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()
	req.RegistrationID = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
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
		fnName = "handler::lecturerDistributions"
		req    = new(entity.UseHRfeeForLecturerReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()
	req.RegistrationID = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	if err := req.Validate(); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
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

func (h *reportHandler) getRelatedRegistrations(c *fiber.Ctx) error {
	var (
		fnName = "handler::getRelatedRegistrations"
		req    = new(entity.GetRelatedRegistrationsReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	req.UserID = l.GetUserId()
	req.RegistrationID = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetRelatedRegistrations(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *reportHandler) getRegistrationListPerLecturer(c *fiber.Ctx) error {
	var (
		fnName = "handler::getRegistrationListPerLecturer"
		req    = new(entity.GetRegistrationListPerLecturerReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()
	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
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
		fnName = "handler::getLecturerWages"
		req    = new(entity.GetLecturersWagesReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()
	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
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
func (h *reportHandler) getExportedLecturersWages(c *fiber.Ctx) error {
	var (
		req    = new(entity.GetExportedLecturersWagesReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
		fnName = "handler::getExportedLecturerWages"
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()
	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetExportedLecturersWages(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	defer func() {
		if err := os.Remove(resp.FilePath); err != nil {
			log.Warn().Err(err).Msgf("%s - delete file error", fnName)
		}
	}()

	if err := c.Download(resp.FilePath, resp.FileName); err != nil {
		log.Warn().Err(err).Msgf("%s - download file error", fnName)
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error(err))
	}

	return nil
}

func (h *reportHandler) importLecturersWages(c *fiber.Ctx) error {
	var (
		req    = new(entity.ImportLecturersWagesReq)
		l      = m.GetLocals(c)
		fnName = "handler::importLecturersWages"
	)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		log.Warn().Err(err).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	file, err := fileHeader.Open()
	if err != nil {
		log.Warn().Err(err).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}
	defer file.Close()

	req.File = file
	req.UserID = l.GetUserId()

	if err := h.service.ImportLecturersWages(c.Context(), req); err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(nil, ""))
}

func (h *reportHandler) getLecturerWagesAggregate(c *fiber.Ctx) error {
	var (
		fnName = "handler::getLecturerWagesAggregate"
		req    = new(entity.GetLecturersWagesAggregateReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()
	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
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

func (h *reportHandler) getLecturerWagesAggregateYearly(c *fiber.Ctx) error {
	var (
		fnName = "handler::getLecturerWagesAggregateYearly"
		req    = new(entity.GetLecturersWagesAggregateYearlyReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()
	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetLecturersWagesAggregateYearly(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *reportHandler) getAcquisitionRightsAggregate(c *fiber.Ctx) error {
	var (
		fnName = "handler::getAcquisitionRightsAggregate"
		req    = new(entity.GetAcquisitionRightsAggregateReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()
	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
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
		fnName = "handler::updateLecturerWages"
		req    = new(entity.UpdateLecturersWageReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()
	req.RegistrationID = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	if err := req.Validate(); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
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
		fnName = "handler::generateRegistrationReports"
		req    = new(entity.GenerateRegistrationsReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.SetDefault()
	req.UserID = l.GetUserId()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
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
		fnName = "handler::registrationMultiAllocation"
		req    = new(entity.RegistrationMuliAllocationReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()
	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
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
