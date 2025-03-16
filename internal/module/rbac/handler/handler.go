package handler

import (
	"codebase-app/internal/adapter"
	m "codebase-app/internal/middleware"
	"codebase-app/internal/module/rbac/entity"
	"codebase-app/internal/module/rbac/ports"
	"codebase-app/internal/module/rbac/repository"
	"codebase-app/internal/module/rbac/service"
	"codebase-app/pkg/errmsg"
	"codebase-app/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
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
	router.Get("/role-access-rights", m.AuthBearer, h.GetRoleAndPermissions)
}

func (h *rbacHandler) GetRoleAndPermissions(c *fiber.Ctx) error {
	var (
		req = new(entity.GetRoleAndPermissionsReq)
		v   = adapter.Adapters.Validator
		l   = m.GetLocals(c)
	)

	req.UserId = l.GetUserId()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::GetRoleAndPermissions - invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetRoleAndPermissions(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp.Items, ""))
}
