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

func (h *payrollHandler) GetPayrollItemsPeriodically(c *fiber.Ctx) error {
	ctx, span := tracing.StartSpan(c.UserContext(), "handler.GetPayrollItemsPeriodically")
	defer span.End()

	var req entity.GetPayrollItemsPeriodicallyReq
	if err := c.QueryParser(&req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("GetPayrollItemsPeriodically - Parse query")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.SetDefault()

	if err := adapter.Adapters.Validator.Validate(&req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("GetPayrollItemsPeriodically - Validate")
		code, errs := errmsg.Errors(err, &req)
		return c.Status(code).JSON(response.Error(errs))
	}

	respData, err := h.service.GetPayrollItemsPeriodically(ctx, &req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(respData, ""))
}

func (h *payrollHandler) GetPayrollReportsYearly(c *fiber.Ctx) error {
	ctx, span := tracing.StartSpan(c.UserContext(), "handler.GetPayrollReportsYearly")
	defer span.End()

	var req entity.GetPayrollReportsYearlyReq
	if err := c.QueryParser(&req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("GetPayrollReportsYearly - Parse query")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.SetDefault()

	if err := adapter.Adapters.Validator.Validate(&req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("GetPayrollReportsYearly - Validate")
		code, errs := errmsg.Errors(err, &req)
		return c.Status(code).JSON(response.Error(errs))
	}

	respData, err := h.service.GetPayrollReportsYearly(ctx, &req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(respData, ""))
}
