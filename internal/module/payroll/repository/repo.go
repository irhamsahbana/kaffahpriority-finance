package repository

import (
	"codebase-app/internal/adapter"
	ports "codebase-app/internal/ports/module/payroll"

	"github.com/jmoiron/sqlx"
)

var _ ports.PayrollRepository = &payrollRepo{}

type payrollRepo struct {
	db *sqlx.DB
}

func NewPayrollRepository() *payrollRepo {
	return &payrollRepo{
		db: adapter.Adapters.Postgres,
	}
}
