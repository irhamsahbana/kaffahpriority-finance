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

func (h *rbacHandler) UpdateRole(c *fiber.Ctx) error {
	var (
		ctx    = c.UserContext()
		fnName = "handler::UpdateRole"
		req    = new(entity.UpdateRoleReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	req.UserID = l.GetUserId()
	req.ID = c.Params("id")

	if err := c.BodyParser(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.UpdateRole(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}
