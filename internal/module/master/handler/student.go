package handler

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/entity"
	m "codebase-app/internal/middleware"
	"codebase-app/pkg/errmsg"
	"codebase-app/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func (h *masterHandler) getStudents(c *fiber.Ctx) error {
	var (
		ctx = c.UserContext()
		req = new(entity.GetStudentsReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	req.UserID = l.GetUserId()

	if err := c.QueryParser(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("failed to parse request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msg("invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetStudents(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(resp, ""))
}

func (h *masterHandler) createStudent(c *fiber.Ctx) error {
	var (
		ctx = c.UserContext()
		req = new(entity.CreateStudentReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	req.UserID = l.GetUserId()

	if err := c.BodyParser(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("failed to parse request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msg("invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.CreateStudent(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success(resp, ""))
}

func (h *masterHandler) getStudent(c *fiber.Ctx) error {
	var (
		ctx = c.UserContext()
		req = new(entity.GetStudentReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	req.ID = c.Params("id")
	req.UserID = l.GetUserId()

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msg("invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetStudent(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(resp, ""))
}

func (h *masterHandler) updateStudent(c *fiber.Ctx) error {
	var (
		ctx = c.UserContext()
		req = new(entity.UpdateStudentReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	req.ID = c.Params("id")
	req.UserID = l.GetUserId()

	if err := c.BodyParser(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("failed to parse request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msg("invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	err := h.service.UpdateStudent(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(nil, ""))
}

func (h *masterHandler) deleteStudent(c *fiber.Ctx) error {
	var (
		ctx = c.UserContext()
		req = new(entity.DeleteStudentReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	req.ID = c.Params("id")
	req.UserID = l.GetUserId()

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msg("invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	err := h.service.DeleteStudent(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(nil, ""))
}
