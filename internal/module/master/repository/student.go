package repository

import (
	"codebase-app/internal/module/master/entity"
	"codebase-app/pkg/errmsg"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (r *masterRepo) GetStudents(ctx context.Context, req *entity.GetStudentsReq) (*entity.GetStudentsResp, error) {
	type dao struct {
		TotalData int `db:"total_data"`
		entity.Student
	}

	var (
		resp       = new(entity.GetStudentsResp)
		data       = make([]dao, 0)
		args       = make([]any, 0, 3)
		StudentIds = make([]string, 0)
	)
	resp.Items = make([]entity.Student, 0)

	query := `
		SELECT
			COUNT (*) OVER() AS total_data,
			id,
			identifier,
			name,
			is_active,
			CASE
				WHEN registered_at IS NULL THEN NULL
				ELSE TO_CHAR(registered_at, 'YYYY-MM-DD')
			END AS registered_at
		FROM
			students
		WHERE
			deleted_at IS NULL
	`

	if req.IsActive != "" {
		query += ` AND is_active = ?`
		args = append(args, req.IsActive)
	}

	if req.Q != "" {
		query += ` AND (
			name ILIKE '%' || ? || '%' OR
			identifier ILIKE '%' || ? || '%'
		)
		`
		args = append(args, req.Q, req.Q)
	}

	query += ` LIMIT ? OFFSET ?`
	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	if err := r.db.SelectContext(ctx, &data,
		r.db.Rebind(query), args...); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetStudents - failed to query students")
		return nil, err
	}

	var lastPaymentAt = make(map[string]*string)

	for _, d := range data {
		resp.Meta.TotalData = d.TotalData
		resp.Items = append(resp.Items, d.Student)
		StudentIds = append(StudentIds, d.Id)

		lastPaymentAt[d.Id] = nil
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	// get last payment at

	if len(StudentIds) > 0 {
		query := `
			SELECT
				student_id,
				MAX(created_at) AS last_payment_at
			FROM
				program_registrations
			WHERE
				student_id IN (?)
				AND deleted_at IS NULL
			GROUP BY
				student_id
		`

		query, args, err := sqlx.In(query, StudentIds)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msg("repo::GetStudents - failed to query last payment at")
			return nil, err
		}

		query = r.db.Rebind(query)

		var lastPaymentAtData = make([]struct {
			StudentId     string  `db:"student_id"`
			LastPaymentAt *string `db:"last_payment_at"`
		}, 0)

		if err := r.db.SelectContext(ctx, &lastPaymentAtData, query, args...); err != nil {
			log.Error().Err(err).Any("req", req).Msg("repo::GetStudents - failed to query last payment at")
			return nil, err
		}

		for _, d := range lastPaymentAtData {
			lastPaymentAt[d.StudentId] = d.LastPaymentAt
		}
	}

	for i, student := range resp.Items {
		if v, ok := lastPaymentAt[student.Id]; ok {
			resp.Items[i].LastPaymentAt = v
		}
	}

	return resp, nil
}

func (r *masterRepo) CreateStudent(ctx context.Context, req *entity.CreateStudentReq) (*entity.CreateStudentResp, error) {
	query := `
		INSERT INTO students (
			id,
			identifier,
			name,
			registered_at,
			is_active
		) VALUES (?, ?, ?, ?, TRUE)
	`

	var (
		Id   = ulid.Make().String()
		resp = new(entity.CreateStudentResp)
	)

	if _, err := r.db.ExecContext(ctx, r.db.Rebind(query),
		Id, req.Identifier, req.Name, req.RegisteredAt); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::CreateStudent - failed to create student")
		return nil, err
	}

	resp.Id = Id

	return resp, nil
}

func (r *masterRepo) GetStudent(ctx context.Context, req *entity.GetStudentReq) (*entity.GetStudentResp, error) {
	var (
		resp = new(entity.GetStudentResp)
		data = new(entity.Student)
	)

	query := `
		SELECT
			id,
			identifier,
			name,
			CASE
				WHEN registered_at IS NULL THEN NULL
				ELSE TO_CHAR(registered_at, 'YYYY-MM-DD')
			END AS registered_at,
			is_active
		FROM
			students
		WHERE
			id = ?
			AND deleted_at IS NULL
	`

	if err := r.db.GetContext(ctx, data, r.db.Rebind(query), req.Id); err != nil {
		if err == sql.ErrNoRows {
			log.Warn().Any("req", req).Msg("repo::GetStudent - student not found")
			return nil, errmsg.NewCustomErrors(404).SetMessage("Santri tidak ditemukan")
		}
		log.Error().Err(err).Any("req", req).Msg("repo::GetStudent - failed to get student")
		return nil, err
	}

	resp.Student = *data

	return resp, nil
}

func (r *masterRepo) UpdateStudent(ctx context.Context, req *entity.UpdateStudentReq) error {
	query := `
		UPDATE students
		SET
			identifier = ?,
			name = ?,
			registered_at = ?,
			is_active = ?,
			updated_at = NOW()
		WHERE
			id = ?
			AND deleted_at IS NULL
	`

	if _, err := r.db.ExecContext(ctx, r.db.Rebind(query),
		req.Identifier, req.Name, req.RegisteredAt, req.IsActive, req.Id); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateStudent - failed to update student")
		return err
	}

	return nil
}

func (r *masterRepo) DeleteStudent(ctx context.Context, req *entity.DeleteStudentReq) error {
	query := `
		UPDATE students
		SET
			deleted_at = NOW()
		WHERE
			id = ?
			AND deleted_at IS NULL
	`

	if _, err := r.db.ExecContext(ctx, r.db.Rebind(query), req.Id); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::DeleteStudent - failed to delete student")
		return err
	}

	return nil
}
