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

func (h *reportHandler) createRegistrations(c *fiber.Ctx) error {
	var (
		ctx, span = tracing.StartSpan(c.UserContext(), "handler.createRegistrations")
		req       = new(entity.CreateRegistrationsReq)
		v         = adapter.Adapters.Validator
		l         = m.GetLocals(c)
	)
	defer span.End()

	if err := c.BodyParser(&req.Registrations); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msg("invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	err := h.service.CreateRegistrations(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success(nil, ""))
}

func (h *reportHandler) copyRegistrations(c *fiber.Ctx) error {
	var (
		ctx, span = tracing.StartSpan(c.UserContext(), "handler.copyRegistrations")
		req       = new(entity.CopyRegistrationsReq)
		v         = adapter.Adapters.Validator
		l         = m.GetLocals(c)
	)
	defer span.End()

	if err := c.BodyParser(&req.Registrations); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msg("invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	err := h.service.CopyRegistrations(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success(nil, ""))
}

func (h *reportHandler) updateRegistration(c *fiber.Ctx) error {
	var (
		ctx, span = tracing.StartSpan(c.UserContext(), "handler.updateRegistration")
		req       = new(entity.UpdateRegistrationReq)
		v         = adapter.Adapters.Validator
		l         = m.GetLocals(c)
	)
	defer span.End()

	if err := c.BodyParser(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()
	req.ID = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msg("invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	if err := req.Validate(); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msg("invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.UpdateRegistration(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *reportHandler) getRegistrations(c *fiber.Ctx) error {
	var (
		ctx, span = tracing.StartSpan(c.UserContext(), "handler.getRegistrations")
		req       = new(entity.GetRegistrationsReq)
		v         = adapter.Adapters.Validator
		l         = m.GetLocals(c)
	)
	defer span.End()

	if err := c.QueryParser(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()
	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msg("invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	if err := req.Validate(); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msg("invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetRegistrations(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *reportHandler) getUnusedRegistrations(c *fiber.Ctx) error {
	var (
		ctx, span = tracing.StartSpan(c.UserContext(), "handler.getUnusedRegistrations")
		req       = new(entity.GetRegistrationsReq)
		v         = adapter.Adapters.Validator
		l         = m.GetLocals(c)
	)
	defer span.End()

	if err := c.QueryParser(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()
	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msg("invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	if err := req.Validate(); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msg("invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetUnusedRegistrations(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *reportHandler) updateRegistrationLecturer(c *fiber.Ctx) error {
	var (
		ctx, span = tracing.StartSpan(c.UserContext(), "handler.updateRegistrationLecturer")
		req       = new(entity.UpdateRegistrationLecturerReq)
		v         = adapter.Adapters.Validator
		l         = m.GetLocals(c)
	)
	defer span.End()

	if err := c.BodyParser(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()
	req.ID = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msg("invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.UpdateRegistrationLecturer(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *reportHandler) updateRegistrationIsPaid(c *fiber.Ctx) error {
	var (
		ctx = c.UserContext()
		req = new(entity.UpdateRegistrationIsPaidReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	if err := c.BodyParser(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()
	req.ID = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msg("invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.UpdateRegistrationIsPaid(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *reportHandler) registrationsMarkAsUsed(c *fiber.Ctx) error {
	var (
		ctx, span = tracing.StartSpan(c.UserContext(), "handler.registrationsMarkAsUsed")
		req       = new(entity.RegistrationsMarkAsUsedReq)
		v         = adapter.Adapters.Validator
		l         = m.GetLocals(c)
	)
	defer span.End()

	if err := c.BodyParser(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = l.GetUserId()

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msg("invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	err := h.service.RegistrationsMarkAsUsed(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(nil, ""))
}
