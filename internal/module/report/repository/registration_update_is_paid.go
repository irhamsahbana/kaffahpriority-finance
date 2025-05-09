package repository

import (
	"codebase-app/internal/module/report/entity"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *reportRepo) UpdateRegistrationIsPaid(ctx context.Context, req *entity.UpdateRegistrationIsPaidReq) (*entity.UpdateRegistrationIsPaidResp, error) {
	var (
		resp = new(entity.UpdateRegistrationIsPaidResp)
	)

	query := `
		UPDATE program_registrations
		SET is_paid = $1
		WHERE id = $2
		AND deleted_at IS NULL
		RETURNING id
	`

	err := r.db.QueryRowxContext(ctx, query, req.IsPaid, req.Id).Scan(&resp.Id)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateRegistrationIsPaid - error updating registration is paid")
		return nil, err
	}

	return resp, nil
}
