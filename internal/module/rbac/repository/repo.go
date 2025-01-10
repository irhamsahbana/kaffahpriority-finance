package repository

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/module/z_template_v2/ports"

	"github.com/jmoiron/sqlx"
)

var _ ports.XxxRepository = &xxxRepo{}

type xxxRepo struct {
	db *sqlx.DB
}

func NewXxxRepository() *xxxRepo {
	return &xxxRepo{
		db: adapter.Adapters.Postgres,
	}
}
