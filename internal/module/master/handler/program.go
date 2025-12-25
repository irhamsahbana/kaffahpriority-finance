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

func (h *masterHandler) getPrograms(c *fiber.Ctx) error {
	var (
		ctx    = c.UserContext()
		fnName = "handler::getPrograms"
		req    = new(entity.GetProgramsReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	req.UserID = l.GetUserId()

	if err := c.QueryParser(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msgf("%s - failed to parse request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetPrograms(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(resp, ""))
}

func (h *masterHandler) getProgram(c *fiber.Ctx) error {
	var (
		ctx    = c.UserContext()
		fnName = "handler::getProgram"
		req    = new(entity.GetProgramReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	req.ID = c.Params("id")
	req.UserID = l.GetUserId()

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetProgram(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(resp, ""))
}

func (h *masterHandler) createProgram(c *fiber.Ctx) error {
	var (
		ctx    = c.UserContext()
		fnName = "handler::createProgram"
		req    = new(entity.CreateProgramReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	req.UserID = l.GetUserId()

	if err := c.BodyParser(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msgf("%s - failed to parse request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.CreateProgram(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success(resp, ""))
}

func (h *masterHandler) updateProgram(c *fiber.Ctx) error {
	var (
		ctx    = c.UserContext()
		fnName = "handler::updateProgram"
		req    = new(entity.UpdateProgramReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	req.ID = c.Params("id")
	req.UserID = l.GetUserId()

	if err := c.BodyParser(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msgf("%s - failed to parse request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.UpdateProgram(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(resp, ""))
}

func (h *masterHandler) deleteProgram(c *fiber.Ctx) error {
	var (
		ctx    = c.UserContext()
		fnName = "handler::deleteProgram"
		req    = new(entity.DeleteProgramReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	req.ID = c.Params("id")
	req.UserID = l.GetUserId()

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	err := h.service.DeleteProgram(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(nil, ""))
}
