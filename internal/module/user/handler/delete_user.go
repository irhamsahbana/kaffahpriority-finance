package handler

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/internal/middleware"
	"codebase-app/pkg/errmsg"
	"codebase-app/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func (h *userHandler) deleteUser(c *fiber.Ctx) error {
	var (
		ctx, span = tracing.StartSpan(c.UserContext(), "handler.deleteUser")
		req       = new(entity.DeleteUserReq)
		v         = adapter.Adapters.Validator
		l         = middleware.GetLocals(c)
	)
	defer span.End()

	req.UserID = l.GetUserId()
	req.ID = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msg("Invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	err := h.service.DeleteUser(ctx, req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(nil, ""))
}
