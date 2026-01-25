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

func (r *masterRepo) GetLecturers(ctx context.Context, req *entity.GetLecturersReq) (*entity.GetLecturersResp, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.GetLecturers")
	defer span.End()

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
			l.id,
			l.academic_manager_id,
			am.name AS academic_manager_name,
			l.name,
			l.phone,
			CASE
				WHEN registered_at IS NOT NULL THEN TO_CHAR(registered_at, 'YYYY-MM-DD')
				ELSE NULL
			END AS registered_at
		FROM
			lecturers l
		LEFT JOIN academic_managers am ON l.academic_manager_id = am.id
		WHERE
			l.deleted_at IS NULL
	`

	if req.Q != "" {
		query += ` AND (
			l.name ILIKE '%' || ? || '%' OR
			l.phone ILIKE '%' || ? || '%'
		)
		`
		args = append(args, req.Q, req.Q)
	}

	query += ` LIMIT ? OFFSET ?`
	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	if err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...); err != nil {
		log.Error().Err(err).Any("req", req).Msg("failed to query lecturers")
		return nil, err
	}

	for _, d := range data {
		resp.Meta.TotalData = d.TotalData
		resp.Items = append(resp.Items, d.Lecturer)
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	return resp, nil
}

func (r *masterRepo) CreateLecturer(ctx context.Context, req *entity.CreateLecturerReq) (*entity.CreateLecturerResp, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.CreateLecturer")
	defer span.End()

	query := `
		INSERT INTO lecturers (
			id,
			academic_manager_id,
			name,
			phone,
			registered_at
		) VALUES (?, ?, ?, ?, ?)
	`

	var (
		Id   = ulid.Make().String()
		resp = new(entity.CreateLecturerResp)
	)

	if _, err := r.db.ExecContext(ctx, r.db.Rebind(query),
		Id, req.AcademicManagerID, req.Name, req.Phone, req.RegisteredAt); err != nil {
		log.Error().Err(err).Any("req", req).Msg("failed to create lecturer")
		return nil, err
	}

	resp.ID = Id

	return resp, nil
}

func (r *masterRepo) GetLecturer(ctx context.Context, req *entity.GetLecturerReq) (*entity.GetLecturerResp, error) {
	var (
		resp = new(entity.GetLecturerResp)
		data = new(entity.Lecturer)
	)

	query := `
		SELECT
			l.id,
			l.academic_manager_id,
			am.name AS academic_manager_name,
			l.name,
			l.phone,
			CASE
				WHEN l.registered_at IS NOT NULL THEN TO_CHAR(l.registered_at, 'YYYY-MM-DD')
				ELSE NULL
			END AS registered_at
		FROM
			lecturers l
		LEFT JOIN academic_managers am ON l.academic_manager_id = am.id
		WHERE
			l.id = ?
			AND l.deleted_at IS NULL
	`

	if err := r.db.GetContext(ctx, data, r.db.Rebind(query), req.ID); err != nil {
		if err == sql.ErrNoRows {
			log.Warn().Any("req", req).Msg("lecturer not found")
			return nil, errmsg.NewCustomErrors(404).SetMessage("Mentor tidak ditemukan")
		}
		log.Error().Err(err).Any("req", req).Msg("failed to get lecturer")
		return nil, err
	}

	resp.Lecturer = *data

	return resp, nil
}

func (r *masterRepo) UpdateLecturer(ctx context.Context, req *entity.UpdateLecturerReq) error {
	query := `
		UPDATE lecturers
		SET
			academic_manager_id = ?,
			name = ?,
			phone = ?,
			registered_at = ?,
			updated_at = NOW()
		WHERE
			id = ?
			AND deleted_at IS NULL
	`

	if _, err := r.db.ExecContext(ctx, r.db.Rebind(query),
		req.AcademicManagerID,
		req.Name, req.Phone, req.RegisteredAt, req.ID); err != nil {
		log.Error().Err(err).Any("req", req).Msg("failed to update lecturer")
		return err
	}

	return nil
}

func (r *masterRepo) DeleteLecturer(ctx context.Context, req *entity.DeleteLecturerReq) error {
	query := `
		UPDATE lecturers
		SET
			deleted_at = NOW()
		WHERE
			id = ?
			AND deleted_at IS NULL
	`

	if _, err := r.db.ExecContext(ctx, r.db.Rebind(query), req.ID); err != nil {
		log.Error().Err(err).Any("req", req).Msg("failed to delete lecturer")
		return err
	}

	return nil
}
