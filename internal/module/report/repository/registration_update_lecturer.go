package repository

import (
	"codebase-app/internal/module/report/entity"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *reportRepo) UpdateRegistrationLecturer(ctx context.Context, req *entity.UpdateRegistrationLecturerReq) (*entity.UpdateRegistrationLecturerResp, error) {
	fnName := "repo::UpdateRegistrationLecturer"
	var (
		resp = new(entity.UpdateRegistrationLecturerResp)
	)

	query := `
		UPDATE program_registrations
		SET lecturer_id = $1
		WHERE id = $2
		AND deleted_at IS NULL
		RETURNING id
	`

	err := r.db.QueryRowxContext(ctx, query, req.LecturerId, req.ID).Scan(&resp.ID)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to update registration lecturer", fnName)
		return nil, err
	}

	return resp, nil
}
