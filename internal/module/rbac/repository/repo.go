package repository

import (
	"codebase-app/internal/adapter"
	ports "codebase-app/internal/ports/module/rbac"

	"github.com/jmoiron/sqlx"
)

var _ ports.RBACRepository = &rbacRepo{}

type rbacRepo struct {
	db *sqlx.DB
}

func NewRBACRepository() *rbacRepo {
	return &rbacRepo{
		db: adapter.Adapters.Postgres,
	}
}
