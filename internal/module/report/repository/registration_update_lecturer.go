package repository

import (
	"codebase-app/internal/module/report/entity"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *reportRepo) UpdateRegistrationLecturer(ctx context.Context, req *entity.UpdateRegistrationLecturerReq) (*entity.UpdateRegistrationLecturerResp, error) {
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

	err := r.db.QueryRowxContext(ctx, query, req.LecturerId, req.Id).Scan(&resp.Id)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateRegistrationLecturer - error updating registration lecturer")
	}

	return resp, nil
}
