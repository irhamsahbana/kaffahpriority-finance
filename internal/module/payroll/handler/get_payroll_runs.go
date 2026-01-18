package handler

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/pkg/errmsg"
	"codebase-app/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func (h *payrollHandler) GetPayrollRuns(c *fiber.Ctx) error {
	ctx, span := tracing.StartSpan(c.UserContext(), "handler.GetPayrollRuns")
	defer span.End()

	var req entity.GetPayrollRunsReq
	if err := c.QueryParser(&req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("GetPayrollRuns - Parse query")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	req.SetDefault()

	resp, err := h.service.GetPayrollRuns(ctx, &req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.JSON(response.Success(resp, ""))
}
