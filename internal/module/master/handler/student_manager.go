package handler

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	m "codebase-app/internal/middleware"
	"codebase-app/pkg/errmsg"
	"codebase-app/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func (h *masterHandler) getStudentManagers(c *fiber.Ctx) error {
	ctx, span := tracing.StartSpan(c.UserContext(), "handler.getStudentManagers")
	defer span.End()

	var (
		req = new(entity.GetStudentManagersReq)
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

	resp, err := h.service.GetStudentManagers(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(resp, ""))
}

func (h *masterHandler) getStudentManager(c *fiber.Ctx) error {
	ctx, span := tracing.StartSpan(c.UserContext(), "handler.getStudentManager")
	defer span.End()

	var (
		req = new(entity.GetStudentManagerReq)
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

	resp, err := h.service.GetStudentManager(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(resp, ""))
}

func (h *masterHandler) createStudentManager(c *fiber.Ctx) error {
	ctx, span := tracing.StartSpan(c.UserContext(), "handler.createStudentManager")
	defer span.End()

	var (
		req = new(entity.CreateStudentManagerReq)
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

	resp, err := h.service.CreateStudentManager(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success(resp, ""))
}

func (h *masterHandler) updateStudentManager(c *fiber.Ctx) error {
	ctx, span := tracing.StartSpan(c.UserContext(), "handler.updateStudentManager")
	defer span.End()

	var (
		req = new(entity.UpdateStudentManagerReq)
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

	err := h.service.UpdateStudentManager(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(nil, ""))
}

func (h *masterHandler) deleteStudentManager(c *fiber.Ctx) error {
	ctx, span := tracing.StartSpan(c.UserContext(), "handler.deleteStudentManager")
	defer span.End()

	var (
		req = new(entity.DeleteStudentManagerReq)
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

	err := h.service.DeleteStudentManager(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(nil, ""))
}

func (h *masterHandler) getAcademicManagers(c *fiber.Ctx) error {
	ctx, span := tracing.StartSpan(c.UserContext(), "handler.getAcademicManagers")
	defer span.End()

	var (
		req = new(entity.GetAcademicManagersReq)
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

	resp, err := h.service.GetAcademicManagers(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(resp, ""))
}

func (h *masterHandler) getAcademicManager(c *fiber.Ctx) error {
	ctx, span := tracing.StartSpan(c.UserContext(), "handler.getAcademicManager")
	defer span.End()

	var (
		req = new(entity.GetAcademicManagerReq)
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

	resp, err := h.service.GetAcademicManager(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(resp, ""))
}

func (h *masterHandler) createAcademicManager(c *fiber.Ctx) error {
	ctx, span := tracing.StartSpan(c.UserContext(), "handler.createAcademicManager")
	defer span.End()

	var (
		req = new(entity.CreateAcademicManagerReq)
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

	resp, err := h.service.CreateAcademicManager(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success(resp, ""))
}

func (h *masterHandler) updateAcademicManager(c *fiber.Ctx) error {
	var (
		ctx = c.UserContext()
		req = new(entity.UpdateAcademicManagerReq)
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

	err := h.service.UpdateAcademicManager(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(nil, ""))
}

func (h *masterHandler) deleteAcademicManager(c *fiber.Ctx) error {
	ctx, span := tracing.StartSpan(c.UserContext(), "handler.deleteAcademicManager")
	defer span.End()

	var (
		req = new(entity.DeleteAcademicManagerReq)
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

	err := h.service.DeleteAcademicManager(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(nil, ""))
}
