package handler

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/pkg/errmsg"
	"codebase-app/pkg/response"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func (h *payrollHandler) ExportPayrollItemsPeriodically(c *fiber.Ctx) error {
	ctx, span := tracing.StartSpan(c.UserContext(), "handler.ExportPayrollItemsPeriodically")
	defer span.End()

	var req entity.ExportPayrollItemsPeriodicallyReq
	if err := c.QueryParser(&req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("ExportPayrollItemsPeriodically - QueryParser")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	if err := adapter.Adapters.Validator.Validate(&req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("ExportPayrollItemsPeriodically - Validate")
		code, errs := errmsg.Errors(err, &req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.ExportPayrollItemsPeriodically(ctx, &req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	defer func() {
		if err := os.Remove(resp.FilePath); err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("ExportPayrollItemsPeriodically - DeleteFile")
		}
	}()

	if err := c.Download(resp.FilePath, resp.FileName); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("ExportPayrollItemsPeriodically - Download")
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error(err))
	}

	return nil
}
