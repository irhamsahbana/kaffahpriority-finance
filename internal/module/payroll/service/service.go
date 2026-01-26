package service

import (
	portsActivityLog "codebase-app/internal/ports/module/activity_log"
	portsMaster "codebase-app/internal/ports/module/master"
	ports "codebase-app/internal/ports/module/payroll"
	portsUser "codebase-app/internal/ports/module/user"
)

var _ ports.PayrollService = &payrollService{}

type payrollService struct {
	repo            ports.PayrollRepository
	activityLogRepo portsActivityLog.ActivityLogRepository
	userRepo        portsUser.UserRepository
	masterRepo      portsMaster.MasterRepository
}

type Config struct {
	Repo            ports.PayrollRepository
	ActivityLogRepo portsActivityLog.ActivityLogRepository
	UserRepo        portsUser.UserRepository
	MasterRepo      portsMaster.MasterRepository
}

func NewPayrollService(cfg Config) *payrollService {
	return &payrollService{
		repo:            cfg.Repo,
		activityLogRepo: cfg.ActivityLogRepo,
		userRepo:        cfg.UserRepo,
		masterRepo:      cfg.MasterRepo,
	}
}
