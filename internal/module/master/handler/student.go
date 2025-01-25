package handler

import (
	"codebase-app/internal/adapter"
	m "codebase-app/internal/middleware"
	"codebase-app/internal/module/master/entity"
	"codebase-app/pkg/errmsg"
	"codebase-app/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func (h *masterHandler) getStudents(c *fiber.Ctx) error {
	var (
		req = new(entity.GetStudentsReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	req.UserId = l.GetUserId()

	if err := c.QueryParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::getStudents - failed to parse request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getStudents - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetStudents(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(resp, ""))
}

func (h *masterHandler) createStudent(c *fiber.Ctx) error {
	var (
		req = new(entity.CreateStudentReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	req.UserId = l.GetUserId()

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::createStudent - failed to parse request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::createStudent - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.CreateStudent(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(resp, ""))
}

func (h *masterHandler) getStudent(c *fiber.Ctx) error {
	var (
		req = new(entity.GetStudentReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	req.Id = c.Params("id")
	req.UserId = l.GetUserId()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getStudent - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetStudent(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(resp, ""))
}

func (h *masterHandler) updateStudent(c *fiber.Ctx) error {
	var (
		req = new(entity.UpdateStudentReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	req.Id = c.Params("id")
	req.UserId = l.GetUserId()

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Msg("handler::updateStudent - failed to parse request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::updateStudent - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	err := h.service.UpdateStudent(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(nil, ""))
}

func (h *masterHandler) deleteStudent(c *fiber.Ctx) error {
	var (
		req = new(entity.DeleteStudentReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	req.Id = c.Params("id")
	req.UserId = l.GetUserId()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::deleteStudent - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	err := h.service.DeleteStudent(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(nil, ""))
}
