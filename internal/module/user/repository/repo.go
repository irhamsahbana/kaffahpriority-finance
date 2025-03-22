package repository

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/module/user/entity"
	"codebase-app/internal/module/user/ports"
	"codebase-app/pkg/errmsg"
	"codebase-app/pkg/jwthandler"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
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

func (r *userRepo) Login(ctx context.Context, req *entity.LoginReq) (*entity.LoginResp, error) {
	type user struct {
		Id       string `db:"id"`
		Email    string `db:"email"`
		Password string `db:"password"`
		Role     string `db:"role"`
	}
	var (
		res    = new(entity.LoginResp)
		result = new(user)
	)

	query := `
		SELECT
			u.id,
			u.email,
			u.password,
			r.name as role
		FROM
			users u
		JOIN
			roles r ON r.id = u.role_id
		WHERE
			u.email = ?
			AND u.deleted_at IS NULL
	`

	err := r.db.GetContext(ctx, result, r.db.Rebind(query), req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warn().Err(err).Any("req", req.Log()).Msg("repo::Login - User not found")
			return nil, errmsg.NewCustomErrors(400).SetMessage("Kredensial yang Anda masukkan salah")
		}
		log.Error().Err(err).Any("req", req.Log()).Msg("repo::Login - Failed to get user")
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(req.Password)); err != nil {
		log.Warn().Err(err).Any("req", req.Log()).Msg("repo::Login - Password not match")
		return nil, errmsg.NewCustomErrors(400).SetMessage("Kredensial yang Anda masukkan salah")
	}

	// set token
	tokenExp := time.Now().UTC().Add(time.Hour * 24)
	payload := jwthandler.CostumClaimsPayload{
		UserId:          result.Id,
		Role:            result.Role,
		TokenExpiration: tokenExp,
	}

	token, err := jwthandler.GenerateTokenString(payload)
	if err != nil {
		log.Error().Err(err).Any("req", req.Log()).Msg("repo::Login - Failed to generate token")
		return nil, errmsg.NewCustomErrors(500).SetMessage("Gagal membuat token")
	}

	res.AccessToken = token

	return res, nil
}

func (r *userRepo) GetUsers(ctx context.Context, req *entity.GetUsersReq) (*entity.GetUsersResp, error) {
	type dao struct {
		TotalData int `db:"total_data"`
		entity.UserItem
	}
	var (
		resp = new(entity.GetUsersResp)
		data = make([]dao, 0)
		args = make([]any, 0)
	)
	resp.Items = make([]entity.UserItem, 0)

	query := `
		SELECT
			COUNT(*) OVER() AS total_data,
			u.id,
			r.id as role_id,
			u.name,
			r.name as role,
			u.email
		FROM
			users u
		JOIN
			roles r ON r.id = u.role_id
		LIMIT ? OFFSET ?
	`

	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetUsers - failed to fetch data")
		return nil, err
	}

	for _, item := range data {
		resp.Meta.TotalData = item.TotalData
		resp.Items = append(resp.Items, item.UserItem)
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	return resp, nil
}

func (r *userRepo) GetUser(ctx context.Context, req *entity.GetUserReq) (*entity.GetUserResp, error) {
	var resp entity.GetUserResp

	query := `
		SELECT
			u.id,
			u.role_id,
			u.name,
			u.email
		FROM
			users u
		WHERE
			u.id = ?
		AND
			u.deleted_at IS NULL
	`

	err := r.db.GetContext(ctx, &resp, r.db.Rebind(query), req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warn().Err(err).Any("req", req).Msg("repo::GetUser - user not found")
			return nil, errmsg.NewCustomErrors(404).SetMessage("Data tidak ditemukan")
		}
		log.Error().Err(err).Any("req", req).Msg("repo::GetUser - failed to fetch user")
		return nil, err
	}

	return &resp, nil
}

func (r *userRepo) UpdateUser(ctx context.Context, req *entity.UpdateUserReq) (*entity.UpdateUserResp, error) {
	var (
		resp       entity.UpdateUserResp
		queryParts = []string{}
		args       = []any{}
	)
	resp.Id = req.Id

	queryParts = append(queryParts, `
		role_id = ?,
		name = ?,
		email = ?
	`)
	args = append(args, req.RoleId, req.Name, req.Email)

	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msg("repo::UpdateUser - failed to hash password")
			return nil, err
		}

		queryParts = append(queryParts, "password = ?")
		args = append(args, hashedPassword)
	}

	query := `
		UPDATE
			users
		SET
			%s
		WHERE
			id = ?
	`
	args = append(args, req.Id)

	queryPartsStr := strings.Join(queryParts, ", ")
	query = fmt.Sprintf(query, queryPartsStr)

	_, err := r.db.ExecContext(ctx, r.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateUser - failed to update user")
		return nil, err
	}

	return &resp, nil
}

func (r *userRepo) DeleteUser(ctx context.Context, req *entity.DeleteUserReq) error {
	query := `
		UPDATE
			users
		SET
			deleted_at = NOW()
		WHERE
			id = ?
		AND
			deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, r.db.Rebind(query), req.Id)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::DeleteUser - failed to delete user")
		return err
	}

	return nil
}

func (r *userRepo) CreateUser(ctx context.Context, req *entity.CreateUserReq) (*entity.CreateUserResp, error) {
	var resp entity.CreateUserResp
	resp.Id = ulid.Make().String()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::CreateUser - failed to hash password")
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

	_, err = r.db.ExecContext(ctx, r.db.Rebind(query), resp.Id, req.RoleId, req.Name, req.Email, hashedPassword)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::CreateUser - failed to create user")
		return nil, err
	}

	return &resp, nil
}
