package handler

import (
	m "codebase-app/internal/middleware"
	"codebase-app/internal/module/rbac/repository"
	"codebase-app/internal/module/rbac/service"
	ports "codebase-app/internal/ports/module/rbac"

	"github.com/gofiber/fiber/v2"
)

type rbacHandler struct {
	service ports.RBACService
}

func NewRBACHandler() *rbacHandler {
	var (
		repo    = repository.NewRBACRepository()
		svc     = service.NewRBACService(repo)
		handler = new(rbacHandler)
	)
	handler.service = svc

	return handler
}

func (h *rbacHandler) Register(router fiber.Router) {
	protected := router.Group("/", m.AuthBearer)
	protected.Get("/role-access-rights", h.GetRoleAndPermissions)

	protected.Post("/roles", h.CreateRole)
	protected.Put("/roles/:id", h.UpdateRole)
	protected.Get("/roles/:id", h.GetRoleDetail)
	protected.Put("roles/:id/permissions", h.UpdateRolePermissions)
	protected.Delete("/roles/:id", h.DeleteRole)

	protected.Get("/permissions", h.GetPermissions)
}
