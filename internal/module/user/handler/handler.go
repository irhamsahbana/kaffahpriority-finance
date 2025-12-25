package handler

import (
	"codebase-app/internal/middleware"
	"codebase-app/internal/module/user/repository"
	"codebase-app/internal/module/user/service"
	ports "codebase-app/internal/ports/module/user"

	"github.com/gofiber/fiber/v2"
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
	router.Post("/login", h.login)

	protected := router.Group("/", middleware.AuthBearer)
	protected.Get("/me", h.me)
	protected.Post("/entities", h.createUser)
	protected.Get("/entities", h.getUsers)
	protected.Get("/entities/:id", h.getUser)
	protected.Put("/entities/:id", h.updateUser)
	protected.Delete("/entities/:id", h.deleteUser)
}
