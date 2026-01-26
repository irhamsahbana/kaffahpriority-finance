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

func (h *payrollHandler) GetPayrollItemDetail(c *fiber.Ctx) error {
	ctx, span := tracing.StartSpan(c.UserContext(), "handler.GetPayrollItemDetail")
	defer span.End()

	var req entity.GetPayrollItemDetailReq
	if err := c.ParamsParser(&req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("GetPayrollItemDetail - Parse params")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	if err := adapter.Adapters.Validator.Validate(&req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("GetPayrollItemDetail - Validate")
		code, errs := errmsg.Errors(err, &req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetPayrollItemDetail(ctx, &req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(resp, ""))
}
