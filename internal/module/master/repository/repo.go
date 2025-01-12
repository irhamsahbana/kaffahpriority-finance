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
		entity.Common
	}

	var (
		resp = new(entity.GetMarketersResp)
		data = make([]dao, 0)
	)
	resp.Items = make([]entity.Common, 0)

	query := `
		SELECT
			COUNT (*) OVER() AS total_data,
			id,
			name
		FROM
			marketers
		WHERE
			deleted_at IS NULL
		LIMIT ? OFFSET ?
	`

	if err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), req.Paginate, (req.Page-1)*req.Paginate); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetMarketers - failed to query marketers")
		return nil, err
	}

	for _, d := range data {
		resp.Meta.TotalData = d.TotalData
		resp.Items = append(resp.Items, d.Common)
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	return resp, nil
}

func (r *masterRepo) GetLecturers(ctx context.Context, req *entity.GetLecturersReq) (*entity.GetLecturersResp, error) {
	type dao struct {
		TotalData int `db:"total_data"`
		entity.Common
	}

	var (
		resp = new(entity.GetLecturersResp)
		data = make([]dao, 0)
	)
	resp.Items = make([]entity.Common, 0)

	query := `
		SELECT
			COUNT (*) OVER() AS total_data,
			id,
			name
		FROM
			lecturers
		WHERE
			deleted_at IS NULL
		LIMIT ? OFFSET ?
	`

	if err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), req.Paginate, (req.Page-1)*req.Paginate); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetLecturers - failed to query lecturers")
		return nil, err
	}

	for _, d := range data {
		resp.Meta.TotalData = d.TotalData
		resp.Items = append(resp.Items, d.Common)
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	return resp, nil
}

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
			name
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

func (r *masterRepo) GetPrograms(ctx context.Context, req *entity.GetProgramsReq) (*entity.GetProgramsResp, error) {
	type dao struct {
		TotalData int `db:"total_data"`
		entity.Program
	}

	var (
		resp = new(entity.GetProgramsResp)
		data = make([]dao, 0)
	)
	resp.Items = make([]entity.Program, 0)

	query := `
		SELECT
			COUNT (*) OVER() AS total_data,
			id,
			name,
			price,
			number_of_meetings
		FROM
			programs
		WHERE
			deleted_at IS NULL
		LIMIT ? OFFSET ?
	`

	if err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), req.Paginate, (req.Page-1)*req.Paginate); err != nil {
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
