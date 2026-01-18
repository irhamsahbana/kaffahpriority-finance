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

func (h *payrollHandler) GetPayrollItems(c *fiber.Ctx) error {
	ctx, span := tracing.StartSpan(c.UserContext(), "handler.GetPayrollItems")
	defer span.End()

	var req entity.GetPayrollItemsReq
	if err := c.QueryParser(&req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("GetPayrollItems - QueryParser")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.SetDefault()

	if err := adapter.Adapters.Validator.Validate(&req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("GetPayrollItems - Validate")
		code, errs := errmsg.Errors(err, &req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetPayrollItems(ctx, &req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(resp, ""))
}
