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

func (h *payrollHandler) GetPayrollRunDetail(c *fiber.Ctx) error {
	ctx, span := tracing.StartSpan(c.UserContext(), "handler.GetPayrollRunDetail")
	defer span.End()

	var req entity.GetPayrollRunDetailReq
	if err := c.ParamsParser(&req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("GetPayrollRunDetail - Parse params")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	if err := adapter.Adapters.Validator.Validate(&req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("GetPayrollRunDetail - Validate")
		code, errs := errmsg.Errors(err, &req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetPayrollRunDetail(ctx, &req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(resp, ""))
}
