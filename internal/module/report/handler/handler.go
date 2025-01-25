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
	router.Put("/templates/:id/generals", m.AuthBearer, h.updateTemplateGeneral)
	router.Put("/templates/:id/finances", m.AuthBearer, h.updateTemplateFinance)
	router.Get("/templates/:id", m.AuthBearer, h.getTemplate)

	router.Post("/registrations", m.AuthBearer, h.createRegistrations)
	router.Post("/copy-registrations", m.AuthBearer, h.copyRegistrations)
	router.Get("/registration-summaries", m.AuthBearer, h.getSummaries)
	router.Get("/registrations", m.AuthBearer, h.getRegistrations)
	router.Put("/registrations/:id", m.AuthBearer, h.updateRegistration)
	router.Get("/registrations/:id", m.AuthBearer, h.getRegistration)

	router.Get("/lecturer-programs", m.AuthBearer, h.getLecturerPrograms)

	// TODO: list of lectures and their programs + students (group by lecture)
}

func (h *reportHandler) getTemplates(c *fiber.Ctx) error {
	var (
		req = new(entity.GetTemplatesReq)
		v   = adapter.Adapters.Validator
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::getTemplates - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getTemplates - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetTemplates(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *reportHandler) getTemplate(c *fiber.Ctx) error {
	var (
		req = new(entity.GetTemplateReq)
		v   = adapter.Adapters.Validator
	)

	req.Id = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getTemplate - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetTemplate(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *reportHandler) createTemplate(c *fiber.Ctx) error {
	var (
		req = new(entity.CreateTemplateReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::createTemplate - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserId = l.GetUserId()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::createTemplate - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	if err := req.Validate(); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::createTemplate - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.CreateTemplate(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success(resp, ""))
}

func (h *reportHandler) updateTemplateGeneral(c *fiber.Ctx) error {
	var (
		req = new(entity.UpdateTemplateGeneralReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::updateTemplate - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserId = l.GetUserId()
	req.Id = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::updateTemplate - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	if err := req.Validate(); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::updateTemplate - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.UpdateTemplateGeneral(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *reportHandler) updateTemplateFinance(c *fiber.Ctx) error {
	var (
		req = new(entity.UpdateTemplateFinanceReq)
		v   = adapter.Adapters.Validator
	)

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::updateTemplate - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.Id = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::updateTemplate - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.UpdateTemplateFinance(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
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

func (h *reportHandler) createRegistrations(c *fiber.Ctx) error {
	var (
		req = new(entity.CreateRegistrationsReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	if err := c.BodyParser(&req.Registrations); err != nil {
		log.Warn().Err(err).Msg("handler::createRegistrations - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserId = l.GetUserId()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::createRegistrations - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	err := h.service.CreateRegistrations(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success(nil, ""))
}

func (h *reportHandler) copyRegistrations(c *fiber.Ctx) error {
	var (
		req = new(entity.CopyRegistrationsReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	if err := c.BodyParser(&req.Registrations); err != nil {
		log.Warn().Err(err).Msg("handler::copyRegistrations - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserId = l.GetUserId()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::copyRegistrations - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	err := h.service.CopyRegistrations(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success(nil, ""))
}

func (h *reportHandler) updateRegistration(c *fiber.Ctx) error {
	var (
		req = new(entity.UpdateRegistrationReq)
		v   = adapter.Adapters.Validator
		// l   = m.GetLocals(c)
	)

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::updateRegistration - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	// req.UserId = l.GetUserId()
	req.Id = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::updateRegistration - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	if err := req.Validate(); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::updateRegistration - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.UpdateRegistration(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *reportHandler) getRegistrations(c *fiber.Ctx) error {
	var (
		req = new(entity.GetRegistrationsReq)
		v   = adapter.Adapters.Validator
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::getRegistrations - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getRegistrations - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	if err := req.Validate(); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getRegistrations - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetRegistrations(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *reportHandler) getRegistration(c *fiber.Ctx) error {
	var (
		req = new(entity.GetRegistrationReq)
		v   = adapter.Adapters.Validator
	)

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

func (h *reportHandler) getLecturerPrograms(c *fiber.Ctx) error {
	var (
		req = new(entity.GetLecturerProgramsReq)
		v   = adapter.Adapters.Validator
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::getLecturerPrograms - invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.SetDefault()

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
