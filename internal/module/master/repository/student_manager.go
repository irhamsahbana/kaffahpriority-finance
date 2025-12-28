package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/pkg/errmsg"
	"context"
	"database/sql"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (r *masterRepo) GetStudentManagers(ctx context.Context, req *entity.GetStudentManagersReq) (*entity.GetStudentManagersResp, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetStudentManagers")
	defer span.End()

	fnName := "repo::GetStudentManagers"
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
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to query student managers", fnName)
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
	ctx, span := tracing.StartSpan(ctx, "repo.CreateStudentManager")
	defer span.End()

	fnName := "repo::CreateStudentManager"
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
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to create student manager", fnName)
		return nil, err
	}

	resp.ID = Id

	return resp, nil
}

func (r *masterRepo) GetStudentManager(ctx context.Context, req *entity.GetStudentManagerReq) (*entity.GetStudentManagerResp, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetStudentManager")
	defer span.End()

	fnName := "repo::GetStudentManager"
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

	if err := r.db.GetContext(ctx, data, r.db.Rebind(query), req.ID); err != nil {
		if err == sql.ErrNoRows {
			log.Warn().Any("req", req).Msgf("%s - student manager not found", fnName)
			return nil, errmsg.NewCustomErrors(404).SetMessage("Pengelola Santri tidak ditemukan")
		}
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to get student manager", fnName)
		return nil, err
	}

	resp.StudentManager = *data

	return resp, nil
}

func (r *masterRepo) UpdateStudentManager(ctx context.Context, req *entity.UpdateStudentManagerReq) error {
	fnName := "repo::UpdateStudentManager"
	query := `
		UPDATE student_managers
		SET
			name = ?,
			updated_at = NOW()
		WHERE
			id = ?
			AND deleted_at IS NULL
	`

	if _, err := r.db.ExecContext(ctx, r.db.Rebind(query), req.Name, req.ID); err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to update student manager", fnName)
		return err
	}

	return nil
}

func (r *masterRepo) DeleteStudentManager(ctx context.Context, req *entity.DeleteStudentManagerReq) error {
	fnName := "repo::DeleteStudentManager"
	query := `
		UPDATE student_managers
		SET
			deleted_at = NOW()
		WHERE
			id = ?
			AND deleted_at IS NULL
	`

	if _, err := r.db.ExecContext(ctx, r.db.Rebind(query), req.ID); err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to delete student manager", fnName)
		return err
	}

	return nil
}
