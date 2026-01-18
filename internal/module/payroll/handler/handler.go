package handler

import (
	m "codebase-app/internal/middleware"
	activityLogRepo "codebase-app/internal/module/activity_log/repository"
	"codebase-app/internal/module/payroll/repository"
	"codebase-app/internal/module/payroll/service"
	userRepo "codebase-app/internal/module/user/repository"
	ports "codebase-app/internal/ports/module/payroll"

	"github.com/gofiber/fiber/v2"
)

type payrollHandler struct {
	service ports.PayrollService
}

func NewPayrollHandler() *payrollHandler {
	repo := repository.NewPayrollRepository()
	repoActivityLog := activityLogRepo.NewActivityLogRepository()
	repoUser := userRepo.NewUserRepository()

	svc := service.NewPayrollService(service.Config{
		Repo:            repo,
		ActivityLogRepo: repoActivityLog,
		UserRepo:        repoUser,
	})
	return &payrollHandler{
		service: svc,
	}
}

func (h *payrollHandler) Register(router fiber.Router) {
	protected := router.Group("/", m.AuthBearer)

	protected.Post("/runs", h.CreatePayrollRun)
	protected.Get("/runs", h.GetPayrollRuns)
	protected.Get("/runs/:id", h.GetPayrollRunDetail)
	protected.Get("/items", h.GetPayrollItems)
	protected.Patch("/items/:id", h.UpdatePayrollItem)
	protected.Delete("/items/:id", h.DeletePayrollItem)
}
