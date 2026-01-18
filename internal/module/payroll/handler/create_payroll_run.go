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

func (h *payrollHandler) CreatePayrollRun(c *fiber.Ctx) error {
	ctx, span := tracing.StartSpan(c.UserContext(), "handler.CreatePayrollRun")
	defer span.End()

	var req entity.CreatePayrollRunReq
	// We ignore BodyParser error because the body is optional now
	_ = c.BodyParser(&req)

	req.UserID, _ = c.Locals("user_id").(string)
	if req.UserID == "" {
		log.Ctx(ctx).Warn().Msg("CreatePayrollRun - Unauthorized")
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("unauthorized"))
	}

	if err := adapter.Adapters.Validator.Validate(&req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("CreatePayrollRun - Validate")
		code, errs := errmsg.Errors(err, &req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.CreatePayrollRun(ctx, &req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(resp, ""))
}
