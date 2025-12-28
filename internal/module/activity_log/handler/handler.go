package handler

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	m "codebase-app/internal/middleware"
	"codebase-app/internal/module/activity_log/repository"
	"codebase-app/internal/module/activity_log/service"
	ports "codebase-app/internal/ports/module/activity_log"
	"codebase-app/pkg/errmsg"
	"codebase-app/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type activityLogHandler struct {
	service ports.ActivityLogService
}

func NewActivityLogHandler() *activityLogHandler {
	var (
		repo    = repository.NewActivityLogRepository()
		svc     = service.NewActivityLogService(repo)
		handler = new(activityLogHandler)
	)
	handler.service = svc

	return handler
}

func (h *activityLogHandler) Register(router fiber.Router) {
	protected := router.Group("/", m.AuthBearer)
	protected.Get("/", h.getActivityLogs)
}

func (h *activityLogHandler) getActivityLogs(c *fiber.Ctx) error {
	var (
		ctx, span = tracing.StartSpan(c.UserContext(), "handler.getActivityLogs")
		fnName    = "handler::getActivityLogs"
		req       = new(entity.GetActivityLogsReq)
		v         = adapter.Adapters.Validator
		l         = m.GetLocals(c)
	)
	defer span.End()

	req.UserID = l.GetUserId()

	if err := c.QueryParser(req); err != nil {
		log.Error().Err(err).Str("fn", fnName).Msg("error parse query")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err))
	}

	if err := v.Validate(req); err != nil {
		log.Ctx(ctx).Warn().Err(err).Any("req", req).Msgf("%s - invalid request", fnName)
		code, errs := errmsg.Errors(err, req)
		return c.Status(code).JSON(response.Error(errs))
	}

	resp, err := h.service.GetActivityLogs(ctx, req)
	if err != nil {
		log.Error().Err(err).Str("fn", fnName).Msg("error get activity logs")
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error(err))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(resp, ""))
}
