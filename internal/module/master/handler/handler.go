package handler

import (
	"codebase-app/internal/adapter"
	m "codebase-app/internal/middleware"
	"codebase-app/internal/module/master/entity"
	"codebase-app/internal/module/master/ports"
	"codebase-app/internal/module/master/repository"
	"codebase-app/internal/module/master/service"
	"codebase-app/pkg/errmsg"
	"codebase-app/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type masterHandler struct {
	service ports.MasterService
}

func NewMasterHandler() *masterHandler {
	var (
		repo    = repository.NewMasterRepository()
		svc     = service.NewMasterService(repo)
		handler = new(masterHandler)
	)
	handler.service = svc

	return handler
}

func (h *masterHandler) Register(router fiber.Router) {
	router.Get("/marketers", m.AuthBearer, h.getMarketers)
	router.Get("/lecturers", m.AuthBearer, h.getLecturers)
	router.Get("/student-managers", m.AuthBearer, h.getStudentManagers)

	router.Get("/students", m.AuthBearer, h.getStudents)
	router.Post("/students", m.AuthBearer, h.createStudent)
	router.Get("/students/:id", m.AuthBearer, h.getStudent)
	router.Put("/students/:id", m.AuthBearer, h.updateStudent)
	router.Delete("/students/:id", m.AuthBearer, h.deleteStudent)

	router.Post("/programs", m.AuthBearer, h.createProgram)
	router.Get("/programs", m.AuthBearer, h.getPrograms)
	router.Get("/programs/:id", m.AuthBearer, h.getProgram)
	router.Put("/programs/:id", m.AuthBearer, h.updateProgram)
	router.Delete("/programs/:id", m.AuthBearer, h.deleteProgram)
}

func (h *masterHandler) getMarketers(c *fiber.Ctx) error {
	var (
		req = new(entity.GetMarketersReq)
		v   = adapter.Adapters.Validator
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::getMarketers - failed to parse request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getMarketers - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetMarketers(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(resp, ""))
}

func (h *masterHandler) getLecturers(c *fiber.Ctx) error {
	var (
		req = new(entity.GetLecturersReq)
		v   = adapter.Adapters.Validator
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::getLecturers - failed to parse request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getLecturers - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetLecturers(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(resp, ""))
}

func (h *masterHandler) getStudentManagers(c *fiber.Ctx) error {
	var (
		req = new(entity.GetStudentManagersReq)
		v   = adapter.Adapters.Validator
	)

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::getStudentManagers - failed to parse request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getStudentManagers - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetStudentManagers(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(resp, ""))
}
