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

func (h *rbacHandler) DeleteRole(c *fiber.Ctx) error {
	var (
		ctx    = c.UserContext()
		fnName = "handler::DeleteRole"
		req    = new(entity.DeleteRoleReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	req.UserID = l.GetUserId()
	req.ID = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	if err := h.service.DeleteRole(ctx, req); err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(nil, ""))
}
