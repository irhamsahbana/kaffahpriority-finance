package repository

import (
	"codebase-app/internal/module/master/entity"
	"context"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (r *masterRepo) GetAcademicManagers(ctx context.Context, req *entity.GetAcademicManagersReq) (*entity.GetAcademicManagersResp, error) {
	type dao struct {
		TotalData int `db:"total_data"`
		entity.AcademicManager
	}

	var (
		resp = new(entity.GetAcademicManagersResp)
		data = make([]dao, 0)
	)
	resp.Items = make([]entity.AcademicManager, 0)

	query := `
		SELECT
			COUNT (*) OVER() AS total_data,
			id,
			name
		FROM
			academic_managers
		WHERE
			deleted_at IS NULL
		LIMIT ? OFFSET ?
	`

	if err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), req.Paginate, (req.Page-1)*req.Paginate); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetAcademicManagers - failed to query academic managers")
		return nil, err
	}

	for _, d := range data {
		resp.Meta.TotalData = d.TotalData
		resp.Items = append(resp.Items, d.AcademicManager)
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	return resp, nil
}

func (r *masterRepo) CreateAcademicManager(ctx context.Context, req *entity.CreateAcademicManagerReq) (*entity.CreateAcademicManagerResp, error) {
	query := `
		INSERT INTO academic_managers (
			id,
			name
		) VALUES (?, ?)
	`

	var (
		id   = ulid.Make().String()
		resp = new(entity.CreateAcademicManagerResp)
	)

	if _, err := r.db.ExecContext(ctx, r.db.Rebind(query), id, req.Name); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::CreateAcademicManager - failed to create academic manager")
		return nil, err
	}

	resp.Id = id

	return resp, nil
}

func (r *masterRepo) GetAcademicManager(ctx context.Context, req *entity.GetAcademicManagerReq) (*entity.GetAcademicManagerResp, error) {
	var (
		resp = new(entity.GetAcademicManagerResp)
		data = new(entity.AcademicManager)
	)

	query := `
		SELECT
			id,
			name
		FROM
			academic_managers
		WHERE
			id = ?
	`

	if err := r.db.GetContext(ctx, data, r.db.Rebind(query), req.Id); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetAcademicManager - failed to query academic manager")
		return nil, err
	}

	resp.AcademicManager = *data

	return resp, nil
}

func (r *masterRepo) UpdateAcademicManager(ctx context.Context, req *entity.UpdateAcademicManagerReq) error {
	query := `
		UPDATE academic_managers
		SET
			name = ?
		WHERE
			id = ?
	`

	if _, err := r.db.ExecContext(ctx, r.db.Rebind(query), req.Name, req.Id); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateAcademicManager - failed to update academic manager")
		return err
	}

	return nil
}

func (r *masterRepo) DeleteAcademicManager(ctx context.Context, req *entity.DeleteAcademicManagerReq) error {
	query := `
		UPDATE academic_managers
		SET
			deleted_at = NOW()
		WHERE
			id = ?
			AND deleted_at IS NULL
	`

	if _, err := r.db.ExecContext(ctx, r.db.Rebind(query), req.Id); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::DeleteAcademicManager - failed to delete academic manager")
		return err
	}

	return nil
}
