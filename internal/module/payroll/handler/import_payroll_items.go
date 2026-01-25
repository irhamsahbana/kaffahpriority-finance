package handler

import (
	"codebase-app/internal/entity"
	"codebase-app/pkg/errmsg"
	"codebase-app/pkg/response"

	"github.com/gofiber/fiber/v2"
)

func (h *payrollHandler) ImportPayrollItems(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("file is required"))
	}

	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("failed to open file"))
	}
	defer src.Close()

	req := &entity.ImportPayrollItemsReq{
		UserID:   c.Locals("user_id").(string),
		File:     src,
		FileName: file.Filename,
	}

	// Validate (custom validation for file presence handled above)
	// We pass the stream to service

	resp, err := h.service.ImportPayrollItems(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, "Import successful"))
}
