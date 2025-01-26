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

	router.Get("/student-managers", m.AuthBearer, h.getStudentManagers)
	router.Get("/student-managers/:id", m.AuthBearer, h.getStudentManager)
	router.Post("/student-managers", m.AuthBearer, h.createStudentManager)
	router.Put("/student-managers/:id", m.AuthBearer, h.updateStudentManager)
	router.Delete("/student-managers/:id", m.AuthBearer, h.deleteStudentManager)

	router.Get("/lecturers", m.AuthBearer, h.getLecturers)
	router.Get("/lecturers/:id", m.AuthBearer, h.getLecturer)
	router.Post("/lecturers", m.AuthBearer, h.createLecturer)
	router.Put("/lecturers/:id", m.AuthBearer, h.updateLecturer)
	router.Delete("/lecturers/:id", m.AuthBearer, h.deleteLecturer)

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
