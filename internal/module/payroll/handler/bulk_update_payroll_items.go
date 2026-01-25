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

func (h *payrollHandler) BulkUpdatePayrollItems(c *fiber.Ctx) error {
	ctx, span := tracing.StartSpan(c.UserContext(), "handler.BulkUpdatePayrollItems")
	defer span.End()

	var req entity.BulkUpdatePayrollItemReq
	if err := c.BodyParser(&req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("BulkUpdatePayrollItems - Parse body")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.UserID, _ = c.Locals("user_id").(string)
	if req.UserID == "" {
		log.Ctx(ctx).Warn().Msg("BulkUpdatePayrollItems - Unauthorized")
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("unauthorized"))
	}

	if err := adapter.Adapters.Validator.Validate(&req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("BulkUpdatePayrollItems - Validate")
		code, errs := errmsg.Errors(err, &req)
		return c.Status(code).JSON(response.Error(errs))
	}

	err := h.service.BulkUpdatePayrollItems(ctx, &req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(nil, "Payroll items updated successfully"))
}
