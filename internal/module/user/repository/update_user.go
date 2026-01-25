package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func (r *userRepo) UpdateUser(ctx context.Context, req *entity.UpdateUserReq) (*entity.UpdateUserResp, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.UpdateUser")
	defer span.End()

	var (
		resp       entity.UpdateUserResp
		queryParts = []string{}
		args       = []any{}
	)
	resp.ID = req.ID

	queryParts = append(queryParts, `
		role_id = ?,
		name = ?,
		email = ?
	`)
	args = append(args, req.RoleID, req.Name, req.Email)

	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msg("failed to hash password")
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
	args = append(args, req.ID)

	queryPartsStr := strings.Join(queryParts, ", ")
	query = fmt.Sprintf(query, queryPartsStr)

	_, err := r.db.ExecContext(ctx, r.db.Rebind(query), args...)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msg("failed to update user")
		return nil, err
	}

	return &resp, nil
}
