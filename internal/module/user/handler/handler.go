package handler

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/module/user/entity"
	"codebase-app/internal/module/user/ports"
	"codebase-app/internal/module/user/repository"
	"codebase-app/internal/module/user/service"
	"codebase-app/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type userHandler struct {
	service ports.UserService
}

func NewUserHandler() *userHandler {
	var (
		repo    = repository.NewUserRepository()
		svc     = service.NewUserService(repo)
		handler = new(userHandler)
	)
	handler.service = svc

	return handler
}

func (h *userHandler) Register(router fiber.Router) {
	v1 := router.Group("/v1")
	v1.Post("/login", h.login)
}

func (h *userHandler) login(c *fiber.Ctx) error {
	var (
		req = new(entity.LoginReq)
		v   = adapter.Adapters.Validator
	)

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Any("req", req.Log()).Msg("handler::login - Invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req.Log()).Msg("handler::login - Invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	resp, err := h.service.Login(c.Context(), req)
	if err != nil {
		log.Error().Err(err).Any("req", req.Log()).Msg("handler::login - Service error")
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error(err))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}
