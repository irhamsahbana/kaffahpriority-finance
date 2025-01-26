package repository

import (
	"codebase-app/internal/module/master/entity"
	"codebase-app/pkg/errmsg"
	"context"
	"database/sql"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

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
			phone,
			email
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

func (r *masterRepo) GetMarketer(ctx context.Context, req *entity.GetMarketerReq) (*entity.GetMarketerResp, error) {
	var (
		resp = new(entity.GetMarketerResp)
		data = new(entity.Marketer)
	)

	query := `
		SELECT
			m.id,
			m.student_manager_id,
			m.name,
			sm.name AS student_manager_name,
			phone,
			email
		FROM
			marketers m
		JOIN
			student_managers sm
			ON m.student_manager_id = sm.id
		WHERE
			m.id = ?
			AND m.deleted_at IS NULL
	`

	if err := r.db.GetContext(ctx, data, r.db.Rebind(query), req.Id); err != nil {
		if err == sql.ErrNoRows {
			log.Warn().Any("req", req).Msg("repo::GetMarketer - marketer not found")
			return nil, errmsg.NewCustomErrors(404).SetMessage("Pemasar tidak ditemukan")
		}
		log.Error().Err(err).Any("req", req).Msg("repo::GetMarketer - failed to get marketer")
		return nil, err
	}

	resp.Marketer = *data

	return resp, nil
}

func (r *masterRepo) CreateMarketer(ctx context.Context, req *entity.CreateMarketerReq) (*entity.CreateMarketerResp, error) {
	query := `
		INSERT INTO marketers (
			id,
			student_manager_id,
			name,
			email,
			phone
		) VALUES (?, ?, ?, ?, ?)
	`

	var (
		Id   = ulid.Make().String()
		resp = new(entity.CreateMarketerResp)
	)

	if _, err := r.db.ExecContext(ctx, r.db.Rebind(query), Id, req.StudentManagerId, req.Name, req.Email, req.Phone); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::CreateMarketer - failed to create marketer")
		return nil, err
	}

	resp.Id = Id

	return resp, nil
}

func (r *masterRepo) UpdateMarketer(ctx context.Context, req *entity.UpdateMarketerReq) error {
	query := `
		UPDATE marketers
		SET
			student_manager_id = ?,
			name = ?,
			email = ?,
			phone = ?,
			updated_at = NOW()
		WHERE
			id = ?
			AND deleted_at IS NULL
	`

	if _, err := r.db.ExecContext(ctx, r.db.Rebind(query), req.StudentManagerId, req.Name, req.Email, req.Phone, req.Id); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateMarketer - failed to update marketer")
		return err
	}

	return nil
}

func (r *masterRepo) DeleteMarketer(ctx context.Context, req *entity.DeleteMarketerReq) error {
	query := `
		UPDATE marketers
		SET
			deleted_at = NOW()
		WHERE
			id = ?
			AND deleted_at IS NULL
	`

	if _, err := r.db.ExecContext(ctx, r.db.Rebind(query), req.Id); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::DeleteMarketer - failed to delete marketer")
		return err
	}

	return nil
}
