package repository

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/module/master/entity"
	"codebase-app/internal/module/master/ports"
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

var _ ports.MasterRepository = &masterRepo{}

type masterRepo struct {
	db *sqlx.DB
}

func NewMasterRepository() *masterRepo {
	return &masterRepo{
		db: adapter.Adapters.Postgres,
	}
}

func (r *masterRepo) GetMarketers(ctx context.Context, req *entity.GetMarketersReq) (*entity.GetMarketersResp, error) {
	type dao struct {
		TotalData int `db:"total_data"`
		entity.Marketer
	}

	var (
		resp = new(entity.GetMarketersResp)
		data = make([]dao, 0)
		args = make([]any, 0, 3)
	)
	resp.Items = make([]entity.Marketer, 0)

	query := `
		SELECT
			COUNT (*) OVER() AS total_data,
			m.id,
			m.student_manager_id,
			m.name,
			sm.name AS student_manager_name,
			phone
		FROM
			marketers m
		JOIN
			student_managers sm
			ON m.student_manager_id = sm.id
		WHERE
			m.deleted_at IS NULL
	`

	if req.Q != "" {
		query += ` AND (
			m.name ILIKE '%' || ? || '%' OR
			m.phone ILIKE '%' || ? || '%'
		)
		`
		args = append(args, req.Q, req.Q)
	}

	query += ` LIMIT ? OFFSET ?`
	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	if err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetMarketers - failed to query marketers")
		return nil, err
	}

	for _, d := range data {
		resp.Meta.TotalData = d.TotalData
		resp.Items = append(resp.Items, d.Marketer)
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	return resp, nil
}

func (r *masterRepo) GetLecturers(ctx context.Context, req *entity.GetLecturersReq) (*entity.GetLecturersResp, error) {
	type dao struct {
		TotalData int `db:"total_data"`
		entity.Lecturer
	}

	var (
		resp = new(entity.GetLecturersResp)
		data = make([]dao, 0)
		args = make([]any, 0, 3)
	)
	resp.Items = make([]entity.Lecturer, 0)

	query := `
		SELECT
			COUNT (*) OVER() AS total_data,
			id,
			name,
			phone
		FROM
			lecturers
		WHERE
			deleted_at IS NULL
	`

	if req.Q != "" {
		query += ` AND (
			name ILIKE '%' || ? || '%' OR
			phone ILIKE '%' || ? || '%'
		)
		`
		args = append(args, req.Q, req.Q)
	}

	query += ` LIMIT ? OFFSET ?`
	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	if err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetLecturers - failed to query lecturers")
		return nil, err
	}

	for _, d := range data {
		resp.Meta.TotalData = d.TotalData
		resp.Items = append(resp.Items, d.Lecturer)
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	return resp, nil
}

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
