package repository

import (
	"codebase-app/internal/module/master/entity"
	"codebase-app/pkg/errmsg"
	"context"
	"database/sql"
	"net/http"

	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (r *masterRepo) GetPrograms(ctx context.Context, req *entity.GetProgramsReq) (*entity.GetProgramsResp, error) {
	type dao struct {
		TotalData int `db:"total_data"`
		entity.Program
	}

	var (
		resp = new(entity.GetProgramsResp)
		data = make([]dao, 0)
		args = make([]any, 0, 3)
	)
	resp.Items = make([]entity.Program, 0)

	query := `
		SELECT
			COUNT (*) OVER() AS total_data,
			id,
			name,
			detail,
			price,
			days,
			lecturer_fee,
			commission_fee,
			price - lecturer_fee - commission_fee AS profit
		FROM
			programs
		WHERE
			deleted_at IS NULL
		`

	if req.Q != "" {
		query += ` AND name ILIKE '%' || ? || '%'`
		args = append(args, req.Q)
	}

	query += `
		LIMIT ? OFFSET ?
	`
	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	if err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetPrograms - failed to query programs")
		return nil, err
	}

	for _, d := range data {
		resp.Meta.TotalData = d.TotalData
		resp.Items = append(resp.Items, d.Program)
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	return resp, nil
}

func (r *masterRepo) GetProgram(ctx context.Context, req *entity.GetProgramReq) (*entity.GetProgramResp, error) {
	var (
		resp = new(entity.GetProgramResp)
	)
	query := `
		SELECT
			id,
			name,
			detail,
			price,
			days,
			lecturer_fee,
			commission_fee,
			price - lecturer_fee - commission_fee AS profit
		FROM
			programs
		WHERE
			id = ?
			AND deleted_at IS NULL
	`

	if err := r.db.GetContext(ctx, resp, r.db.Rebind(query), req.Id); err != nil {
		if err == sql.ErrNoRows {
			log.Warn().Err(err).Any("req", req).Msg("repo::GetProgram - program not found")
			return nil, errmsg.NewCustomErrors(404).SetMessage("Program tidak ditemukan")
		}
		log.Error().Err(err).Any("req", req).Msg("repo::GetProgram - failed to query program")
		return nil, err
	}

	return resp, nil
}

func (r *masterRepo) CreateProgram(ctx context.Context, req *entity.CreateProgramReq) (*entity.CreateProgramResp, error) {
	var (
		resp    = new(entity.CreateProgramResp)
		id      = ulid.Make().String()
		isExist bool
	)

	queryExist := `
		SELECT EXISTS (
			SELECT 1
			FROM programs
			WHERE UPPER(name) = UPPER(?)
			AND deleted_at IS NULL
		)
	`

	if err := r.db.GetContext(ctx, &isExist, r.db.Rebind(queryExist), req.Name); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::CreateProgram - failed to check program exist")
		return nil, err
	}

	if isExist {
		log.Warn().Any("req", req).Msg("repo::CreateProgram - program already exist")
		return nil, errmsg.NewCustomErrors(http.StatusConflict).SetMessage("Nama program sudah ada")
	}

	query := `
		INSERT INTO programs (
			id,
			name,
			detail,
			price,
			days,
			lecturer_fee,
			commission_fee
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, r.db.Rebind(query),
		id,
		req.Name,
		req.Detail,
		req.Price,
		pq.Array(req.Days),
		req.LecturerFee,
		req.CommissionFee,
	)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::CreateProgram - failed to insert program")
		return nil, err
	}

	resp.Id = id

	return resp, nil
}

func (r *masterRepo) UpdateProgram(ctx context.Context, req *entity.UpdateProgramReq) (*entity.UpdateProgramResp, error) {
	var (
		resp    = new(entity.UpdateProgramResp)
		isExist bool
	)

	queryExist := `
		SELECT EXISTS (
			SELECT 1
			FROM programs
			WHERE id = ?
			AND deleted_at IS NULL
		)
	`

	if err := r.db.GetContext(ctx, &isExist, r.db.Rebind(queryExist), req.Id); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateProgram - failed to check program exist")
		return nil, err
	}

	if !isExist {
		log.Warn().Any("req", req).Msg("repo::UpdateProgram - program not found")
		return nil, errmsg.NewCustomErrors(404).SetMessage("Program tidak ditemukan")
	}

	query := `
		UPDATE programs
		SET
			name = ?,
			detail = ?,
			price = ?,
			days = ?,
			lecturer_fee = ?,
			commission_fee = ?
		WHERE
			id = ?
	`

	_, err := r.db.ExecContext(ctx, r.db.Rebind(query),
		req.Name,
		req.Detail,
		req.Price,
		pq.Array(req.Days),
		req.LecturerFee,
		req.CommissionFee,
		req.Id,
	)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateProgram - failed to update program")
		return nil, err
	}

	resp.Id = req.Id

	return resp, nil
}

func (r *masterRepo) DeleteProgram(ctx context.Context, req *entity.DeleteProgramReq) error {
	query := `
		UPDATE programs
		SET
			deleted_at = NOW()
		WHERE
			id = ?
			AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, r.db.Rebind(query), req.Id)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::DeleteProgram - failed to delete program")
		return err
	}

	return nil
}
