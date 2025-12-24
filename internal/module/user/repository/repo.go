package repository

import (
	"codebase-app/internal/adapter"
	ports "codebase-app/internal/ports/module/user"

	"github.com/jmoiron/sqlx"
)

var _ ports.UserRepository = &userRepo{}

type userRepo struct {
	db *sqlx.DB
}

func NewUserRepository() *userRepo {
	return &userRepo{
		db: adapter.Adapters.Postgres,
	}
}
