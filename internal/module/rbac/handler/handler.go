package handler

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/entity"
	m "codebase-app/internal/middleware"
	"codebase-app/internal/module/rbac/repository"
	"codebase-app/internal/module/rbac/service"
	ports "codebase-app/internal/ports/module/rbac"
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

	router.Post("/roles", m.AuthBearer, h.CreateRole)
	router.Put("/roles/:id", m.AuthBearer, h.UpdateRole)
	router.Get("/roles/:id", m.AuthBearer, h.GetRoleDetail)
	router.Put("roles/:id/permissions", m.AuthBearer, h.UpdateRolePermissions)
	router.Delete("/roles/:id", m.AuthBearer, h.DeleteRole)

	router.Get("/permissions", m.AuthBearer, h.GetPermissions)
}

func (h *rbacHandler) GetRoleAndPermissions(c *fiber.Ctx) error {
	var (
		fnName = "handler::GetRoleAndPermissions"
		req    = new(entity.GetRoleAndPermissionsReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	req.UserID = l.GetUserId()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
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

func (h *rbacHandler) CreateRole(c *fiber.Ctx) error {
	var (
		fnName = "handler::CreateRole"
		req    = new(entity.CreateRoleReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	req.UserID = l.GetUserId()

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.CreateRole(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success(resp, ""))
}

func (h *rbacHandler) UpdateRole(c *fiber.Ctx) error {
	var (
		fnName = "handler::UpdateRole"
		req    = new(entity.UpdateRoleReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	req.UserID = l.GetUserId()
	req.ID = c.Params("id")

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.UpdateRole(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *rbacHandler) GetRoleDetail(c *fiber.Ctx) error {
	var (
		fnName = "handler::GetRoleDetail"
		req    = new(entity.GetRoleDetailReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	req.UserID = l.GetUserId()
	req.RoleID = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetRoleDetail(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *rbacHandler) DeleteRole(c *fiber.Ctx) error {
	var (
		fnName = "handler::DeleteRole"
		req    = new(entity.DeleteRoleReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	req.UserID = l.GetUserId()
	req.ID = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	if err := h.service.DeleteRole(c.Context(), req); err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(nil, ""))
}

func (h *rbacHandler) GetPermissions(c *fiber.Ctx) error {
	var (
		fnName = "handler::GetPermissions"
		req    = new(entity.GetPermissionsReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	req.UserID = l.GetUserId()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetPermissions(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp.Items, ""))
}

func (h *rbacHandler) UpdateRolePermissions(c *fiber.Ctx) error {
	var (
		fnName = "handler::UpdateRolePermissions"
		req    = new(entity.UpdateRolePermissionsReq)
		v      = adapter.Adapters.Validator
		l      = m.GetLocals(c)
	)

	req.UserID = l.GetUserId()
	req.RoleID = c.Params("id")

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.UpdateRolePermissions(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}
