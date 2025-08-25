package repository

import (
	"codebase-app/internal/module/report/entity"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *reportRepo) UpdateRegistrationIsPaid(ctx context.Context, req *entity.UpdateRegistrationIsPaidReq) (*entity.UpdateRegistrationIsPaidResp, error) {
	fnName := "repo::UpdateRegistrationIsPaid"
	var (
		resp = new(entity.UpdateRegistrationIsPaidResp)
	)

	query := `
		UPDATE program_registrations
		SET
			is_paid = $1,
			paid_at = CASE WHEN $1 THEN NOW() ELSE paid_at END
		WHERE id = $2
		AND deleted_at IS NULL
		RETURNING id
	`

	err := r.db.QueryRowxContext(ctx, query, req.IsPaid, req.ID).Scan(&resp.ID)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to update registration is paid", fnName)
		return nil, err
	}

	return resp, nil
}
