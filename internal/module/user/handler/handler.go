package handler

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/middleware"
	"codebase-app/internal/module/user/entity"
	"codebase-app/internal/module/user/ports"
	"codebase-app/internal/module/user/repository"
	"codebase-app/internal/module/user/service"
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
	router.Post("/entities", middleware.AuthBearer, h.createUser)
	router.Get("/entities", middleware.AuthBearer, h.getUsers)
	router.Get("/entities/:id", middleware.AuthBearer, h.getUser)
	router.Put("/entities/:id", middleware.AuthBearer, h.updateUser)
	router.Delete("/entities/:id", middleware.AuthBearer, h.deleteUser)
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

func (h *userHandler) getUsers(c *fiber.Ctx) error {
	var (
		req = new(entity.GetUsersReq)
		v   = adapter.Adapters.Validator
		l   = middleware.GetLocals(c)
	)

	req.UserId = l.GetUserId()
	req.SetDefault()

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getUsers - Invalid request")
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
		req = new(entity.GetUserReq)
		v   = adapter.Adapters.Validator
		l   = middleware.GetLocals(c)
	)

	req.UserId = l.GetUserId()
	req.Id = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::getUser - Invalid request")
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
		req = new(entity.UpdateUserReq)
		v   = adapter.Adapters.Validator
		l   = middleware.GetLocals(c)
	)

	req.UserId = l.GetUserId()
	req.Id = c.Params("id")

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::updateUser - Invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::updateUser - Invalid request")
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
		req = new(entity.DeleteUserReq)
		v   = adapter.Adapters.Validator
		l   = middleware.GetLocals(c)
	)

	req.UserId = l.GetUserId()
	req.Id = c.Params("id")

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::deleteUser - Invalid request")
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	err := h.service.DeleteUser(c.Context(), req)
	if err != nil {
		code, errs := errmsg.Errors[error](err)
		return c.Status(code).JSON(response.Error(errs))
	}

	return c.Status(fiber.StatusNoContent).JSON(nil)
}

func (h *userHandler) createUser(c *fiber.Ctx) error {
	var (
		req = new(entity.CreateUserReq)
		v   = adapter.Adapters.Validator
		l   = middleware.GetLocals(c)
	)

	req.UserId = l.GetUserId()

	if err := c.BodyParser(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::createUser - Invalid request")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	if err := v.Validate(req); err != nil {
		log.Warn().Err(err).Any("req", req).Msg("handler::createUser - Invalid request")
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
