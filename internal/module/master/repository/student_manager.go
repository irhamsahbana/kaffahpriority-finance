package repository

import (
	"codebase-app/internal/module/master/entity"
	"codebase-app/pkg/errmsg"
	"context"
	"database/sql"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (r *masterRepo) GetStudentManagers(ctx context.Context, req *entity.GetStudentManagersReq) (*entity.GetStudentManagersResp, error) {
	type dao struct {
		TotalData int `db:"total_data"`
		entity.StudentManager
	}

	var (
		resp = new(entity.GetStudentManagersResp)
		data = make([]dao, 0)
	)
	resp.Items = make([]entity.StudentManager, 0)

	query := `
		SELECT
			COUNT (*) OVER() AS total_data,
			id,
			name
		FROM
			student_managers
		WHERE
			deleted_at IS NULL
		LIMIT ? OFFSET ?
	`

	if err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), req.Paginate, (req.Page-1)*req.Paginate); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetStudentManagers - failed to query student managers")
		return nil, err
	}

	for _, d := range data {
		resp.Meta.TotalData = d.TotalData
		resp.Items = append(resp.Items, d.StudentManager)
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	return resp, nil
}

func (r *masterRepo) CreateStudentManager(ctx context.Context, req *entity.CreateStudentManagerReq) (*entity.CreateStudentManagerResp, error) {
	query := `
		INSERT INTO student_managers (
			id,
			name
		) VALUES (?, ?)
	`

	var (
		Id   = ulid.Make().String()
		resp = new(entity.CreateStudentManagerResp)
	)

	if _, err := r.db.ExecContext(ctx, r.db.Rebind(query), Id, req.Name); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::CreateStudentManager - failed to create student manager")
		return nil, err
	}

	resp.Id = Id

	return resp, nil
}

func (r *masterRepo) GetStudentManager(ctx context.Context, req *entity.GetStudentManagerReq) (*entity.GetStudentManagerResp, error) {
	var (
		resp = new(entity.GetStudentManagerResp)
		data = new(entity.StudentManager)
	)

	query := `
		SELECT
			id,
			name
		FROM
			student_managers
		WHERE
			id = ?
			AND deleted_at IS NULL
	`

	if err := r.db.GetContext(ctx, data, r.db.Rebind(query), req.Id); err != nil {
		if err == sql.ErrNoRows {
			log.Warn().Any("req", req).Msg("repo::GetStudentManager - student manager not found")
			return nil, errmsg.NewCustomErrors(404).SetMessage("Pengelola Santri tidak ditemukan")
		}
		log.Error().Err(err).Any("req", req).Msg("repo::GetStudentManager - failed to get student manager")
		return nil, err
	}

	resp.StudentManager = *data

	return resp, nil
}

func (r *masterRepo) UpdateStudentManager(ctx context.Context, req *entity.UpdateStudentManagerReq) error {
	query := `
		UPDATE student_managers
		SET
			name = ?,
			updated_at = NOW()
		WHERE
			id = ?
			AND deleted_at IS NULL
	`

	if _, err := r.db.ExecContext(ctx, r.db.Rebind(query), req.Name, req.Id); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateStudentManager - failed to update student manager")
		return err
	}

	return nil
}

func (r *masterRepo) DeleteStudentManager(ctx context.Context, req *entity.DeleteStudentManagerReq) error {
	query := `
		UPDATE student_managers
		SET
			deleted_at = NOW()
		WHERE
			id = ?
			AND deleted_at IS NULL
	`

	if _, err := r.db.ExecContext(ctx, r.db.Rebind(query), req.Id); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::DeleteStudentManager - failed to delete student manager")
		return err
	}

	return nil
}
