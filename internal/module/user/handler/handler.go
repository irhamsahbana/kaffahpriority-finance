package handler

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/entity"
	"codebase-app/internal/middleware"
	"codebase-app/internal/module/user/repository"
	"codebase-app/internal/module/user/service"
	ports "codebase-app/internal/ports/module/user"
	"codebase-app/pkg/errmsg"
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
	router.Post("/login", h.login)
	router.Get("/me", middleware.AuthBearer, h.me)
	router.Post("/entities", middleware.AuthBearer, h.createUser)
	router.Get("/entities", middleware.AuthBearer, h.getUsers)
	router.Get("/entities/:id", middleware.AuthBearer, h.getUser)
	router.Put("/entities/:id", middleware.AuthBearer, h.updateUser)
	router.Delete("/entities/:id", middleware.AuthBearer, h.deleteUser)
}

func (h *userHandler) login(c *fiber.Ctx) error {
	var (
		fnName = "handler::login"
		req    = new(entity.LoginReq)
		v      = adapter.Adapters.Validator
	)

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Any("req", req.Log()).Msgf("%s - Invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req.Log()).Msgf("%s - Invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	resp, err := h.service.Login(c.Context(), req)
	if err != nil {
		log.Error().Err(err).Any("req", req.Log()).Msgf("%s - Service error", fnName)
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error(err))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *userHandler) me(c *fiber.Ctx) error {
	var (
		fnName = "handler::me"
		req    = new(entity.GetMeReq)
		v      = adapter.Adapters.Validator
		l      = middleware.GetLocals(c)
	)

	req.UserID = l.GetUserId()
	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - Invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetMe(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *userHandler) getUsers(c *fiber.Ctx) error {
	var (
		fnName = "handler::getUsers"
		req    = new(entity.GetUsersReq)
		v      = adapter.Adapters.Validator
		l      = middleware.GetLocals(c)
	)

	req.UserID = l.GetUserId()
	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - Invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetUsers(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *userHandler) getUser(c *fiber.Ctx) error {
	var (
		fnName = "handler::getUser"
		req    = new(entity.GetUserReq)
		v      = adapter.Adapters.Validator
		l      = middleware.GetLocals(c)
	)

	req.UserID = l.GetUserId()
	req.ID = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - Invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetUser(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *userHandler) updateUser(c *fiber.Ctx) error {
	var (
		fnName = "handler::updateUser"
		req    = new(entity.UpdateUserReq)
		v      = adapter.Adapters.Validator
		l      = middleware.GetLocals(c)
	)

	req.UserID = l.GetUserId()
	req.ID = c.Params("id")

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - Invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - Invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.UpdateUser(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}

func (h *userHandler) deleteUser(c *fiber.Ctx) error {
	var (
		fnName = "handler::deleteUser"
		req    = new(entity.DeleteUserReq)
		v      = adapter.Adapters.Validator
		l      = middleware.GetLocals(c)
	)

	req.UserID = l.GetUserId()
	req.ID = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - Invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	err := h.service.DeleteUser(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(nil, ""))
}

func (h *userHandler) createUser(c *fiber.Ctx) error {
	var (
		fnName = "handler::createUser"
		req    = new(entity.CreateUserReq)
		v      = adapter.Adapters.Validator
		l      = middleware.GetLocals(c)
	)

	req.UserID = l.GetUserId()

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - Invalid request", fnName)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msgf("%s - Invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.CreateUser(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success(resp, ""))
}
