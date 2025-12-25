package handler

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/entity"
	"codebase-app/internal/middleware"
	"codebase-app/pkg/errmsg"
	"codebase-app/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func (h *userHandler) getUsers(c *fiber.Ctx) error {
	var (
		ctx    = c.UserContext()
		fnName = "handler::getUsers"
		req    = new(entity.GetUsersReq)
		v      = adapter.Adapters.Validator
		l      = middleware.GetLocals(c)
	)

	req.UserID = l.GetUserId()
	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msgf("%s - Invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetUsers(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}
