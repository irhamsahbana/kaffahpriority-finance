package repository

import (
	"codebase-app/internal/entity"
	"context"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func (r *userRepo) CreateUser(ctx context.Context, req *entity.CreateUserReq) (*entity.CreateUserResp, error) {
	fnName := "repo::CreateUser"
	var resp entity.CreateUserResp
	resp.ID = ulid.Make().String()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to hash password", fnName)
		return nil, err
	}

	query := `
		INSERT INTO
			users
		(
			id,
			role_id,
			name,
			email,
			password
		) VALUES (
			?, ?, ?, ?, ?
		)
	`

	_, err = r.db.ExecContext(ctx, r.db.Rebind(query), resp.ID, req.RoleID, req.Name, req.Email, hashedPassword)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to create user", fnName)
		return nil, err
	}

	return &resp, nil
}
