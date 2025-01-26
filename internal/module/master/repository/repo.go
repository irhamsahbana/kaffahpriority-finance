package repository

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/module/master/ports"

	"github.com/jmoiron/sqlx"
)

var _ ports.MasterRepository = &masterRepo{}

type masterRepo struct {
	db *sqlx.DB
}

func NewMasterRepository() *masterRepo {
	return &masterRepo{
		db: adapter.Adapters.Postgres,
	}
}
