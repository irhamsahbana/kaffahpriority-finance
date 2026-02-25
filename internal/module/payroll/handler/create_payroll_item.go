package handler

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/pkg/errmsg"
	"codebase-app/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func (h *payrollHandler) CreatePayrollItem(c *fiber.Ctx) error {
	ctx, span := tracing.StartSpan(c.UserContext(), "handler.CreatePayrollItem")
	defer span.End()

	var req entity.CreatePayrollItemReq
	if err := c.BodyParser(&req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("CreatePayrollItem - BodyParser")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID = c.Locals("user_id").(string)

	if err := adapter.Adapters.Validator.Validate(&req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("CreatePayrollItem - Validate")
		code, errs := errmsg.Errors(err, &req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.CreatePayrollItem(ctx, &req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(resp, ""))
}
