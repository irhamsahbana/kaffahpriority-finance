package handler

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/entity"
	"codebase-app/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func (h *userHandler) login(c *fiber.Ctx) error {
	var (
		ctx    = c.UserContext()
		fnName = "handler::login"
		req    = new(entity.LoginReq)
		v      = adapter.Adapters.Validator
	)

	if err := c.BodyParser(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req.Log()).Msgf("%s - Invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req.Log()).Msgf("%s - Invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	resp, err := h.service.Login(ctx, req)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req.Log()).Msgf("%s - Service error", fnName)
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error(err))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}
