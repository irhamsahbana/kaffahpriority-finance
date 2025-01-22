package repository

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/module/report/entity"
	"codebase-app/internal/module/report/ports"
	"codebase-app/pkg/errmsg"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

var _ ports.ReportRepository = &reportRepo{}

type reportRepo struct {
	db *sqlx.DB
}

func NewReportRepository() *reportRepo {
	return &reportRepo{
		db: adapter.Adapters.Postgres,
	}
}

func (r *reportRepo) CreateTemplate(ctx context.Context, req *entity.CreateTemplateReq) (*entity.CreateTemplateResp, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::CreateTemplate - failed to begin transaction")
		return nil, err
	}
	defer func() {
		if err != nil {
			errRB := tx.Rollback()
			if errRB != nil {
				log.Error().Err(errRB).Any("req", req).Msg("repo::CreateTemplate - failed to rollback transaction")
			}
			return
		}
		errCommit := tx.Commit()
		if errCommit != nil {
			log.Error().Err(errCommit).Any("req", req).Msg("repo::CreateTemplate - failed to commit transaction")
		}
	}()

	var (
		Id   = ulid.Make().String()
		resp = new(entity.CreateTemplateResp)
	)
	resp.Id = Id

	query := `
		INSERT INTO program_registration_templates (
			id,
			user_id,
			program_id,
			lecturer_id,
			marketer_id,
			student_id,
			days
		) VALUES (
			?, ?, ?, ?, ?, ?, ?
		)
	`

	_, err = tx.ExecContext(ctx, tx.Rebind(query),
		Id, req.UserId, req.ProgramId, req.LecturerId, req.MarketerId, req.StudentId,
		pq.Array(req.Days),
	)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::CreateTemplate - failed to insert data")
		return nil, err
	}

	for _, item := range req.AdditionalStudents {
		query = `
			INSERT INTO prt_additional_students (
				id, prt_id, student_id, name
			) VALUES (?, ?, ?, ?)
		`

		_, err = tx.ExecContext(ctx, tx.Rebind(query),
			ulid.Make().String(), Id, item.StudentId, item.Name,
		)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msg("repo::CreateTemplate - failed to insert additional students")
			return nil, err
		}
	}

	return resp, nil
}

func (r *reportRepo) GetRegistrations(ctx context.Context, req *entity.GetRegistrationsReq) (*entity.GetRegistrationsResp, error) {
	type dao struct {
		TotalData int `db:"total_data"`
		entity.RegisItem
	}
	var (
		data            = make([]dao, 0, req.Paginate)
		registrationIds = make([]string, 0)
		resp            = new(entity.GetRegistrationsResp)
	)
	resp.Items = make([]entity.RegisItem, 0)

	query := `
		SELECT
			COUNT(*) OVER() AS total_data,
			pr.id,
			pr.program_id,
			pr.marketer_id,
			pr.lecturer_id,
			pr.student_id,
			pr.program_name,
			pr.program_fee,
			pr.administration_fee,
			pr.foreign_learning_fee,
			pr.night_learning_fee,
			pr.marketer_commission_fee,
			pr.overpayment_fee,
			pr.hr_fee,
			pr.marketer_gifts_fee,
			pr.closing_fee_for_office,
			pr.closing_fee_for_reward,
			pr.created_at,
			pr.updated_at,
			pr.program_fee +
			COALESCE(pr.foreign_learning_fee, 0) +
			COALESCE(pr.night_learning_fee, 0) +
			COALESCE(pr.overpayment_fee, 0)
			AS monthly_fee,
			l.name AS lecturer_name,
			m.name AS marketer_name,
			s.name AS student_name
		FROM
			program_registrations pr
		JOIN
			lecturers l
			ON pr.lecturer_id = l.id
		JOIN
			marketers m
			ON pr.marketer_id = m.id
		JOIN
			students s
			ON pr.student_id = s.id
		JOIN
			programs p
			ON pr.program_id = p.id
		WHERE
			pr.deleted_at IS NULL
	`

	err := r.db.SelectContext(ctx, &data, r.db.Rebind(query))
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetRegistrations - failed to fetch data")
		return nil, err
	}

	for _, item := range data {
		resp.Meta.TotalData = item.TotalData
		registrationIds = append(registrationIds, item.Id)
		resp.Items = append(resp.Items, item.RegisItem)
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	if len(registrationIds) > 0 {
		type daos struct {
			PrId string `db:"pr_id"`
			entity.AddStudent
		}
		var (
			daosData = make([]daos, 0)
		)
		query = `
			SELECT
				prs.pr_id,
				prs.student_id,
				CASE
					WHEN s.id IS NULL THEN prs.name
					ELSE s.name
				END AS name
			FROM
				pr_additional_students prs
			LEFT JOIN
				students s
				ON prs.student_id = s.id
			WHERE prs.pr_id IN (?)
		`

		query, args, err := sqlx.In(query, registrationIds)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msg("repo::GetRegistrations - failed to build query")
			return nil, err
		}

		query = r.db.Rebind(query)
		err = r.db.SelectContext(ctx, &daosData, query, args...)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msg("repo::GetRegistrations - failed to fetch additional students")
			return nil, err
		}

		for i, item := range resp.Items {
			for _, data := range daosData {
				if item.Id == data.PrId {
					resp.Items[i].Students = append(resp.Items[i].Students, data.AddStudent)
				}
			}
		}
	}

	return resp, nil
}

func (r *reportRepo) GetRegistration(ctx context.Context, req *entity.GetRegistrationReq) (*entity.GetRegistrationResp, error) {
	var (
		resp = new(entity.GetRegistrationResp)
	)
	resp.Students = make([]entity.AddStudent, 0)

	query := `
		SELECT
			pr.id,
			pr.program_id,
			pr.marketer_id,
			pr.lecturer_id,
			pr.student_id,
			pr.program_name,
			pr.program_fee,
			pr.administration_fee,
			pr.foreign_learning_fee,
			pr.night_learning_fee,
			pr.marketer_commission_fee,
			pr.overpayment_fee,
			pr.hr_fee,
			pr.marketer_gifts_fee,
			pr.closing_fee_for_office,
			pr.closing_fee_for_reward,
			pr.created_at,
			pr.updated_at,
			l.name AS lecturer_name,
			m.name AS marketer_name,
			s.name AS student_name,
			p.name AS program_name
		FROM
			program_registrations pr
		JOIN
			lecturers l
			ON pr.lecturer_id = l.id
		JOIN
			marketers m
			ON pr.marketer_id = m.id
		JOIN
			students s
			ON pr.student_id = s.id
		JOIN
			programs p
			ON pr.program_id = p.id
		WHERE
			pr.id = ?
			AND pr.deleted_at IS NULL
	`

	err := r.db.GetContext(ctx, resp, r.db.Rebind(query), req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warn().Err(err).Any("req", req).Msg("repo::GetRegistration - data not found")
			return nil, errmsg.NewCustomErrors(404).SetMessage("Registrasi tidak ditemukan")
		}
		log.Error().Err(err).Any("req", req).Msg("repo::GetRegistration - failed to fetch data")
		return nil, err
	}

	query = `
		SELECT
			prs.student_id,
			CASE
				WHEN s.id IS NULL THEN prs.name
				ELSE s.name
			END AS name
		FROM
			pr_additional_students prs
		LEFT JOIN
			students s
			ON prs.student_id = s.id
		WHERE
			prs.pr_id = ?
	`

	err = r.db.SelectContext(ctx, &resp.Students, r.db.Rebind(query), req.Id)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetRegistration - failed to fetch additional students")
		return nil, err
	}

	return resp, nil
}

func (r *reportRepo) GetLecturerPrograms(ctx context.Context, req *entity.GetLecturerProgramsReq) (*entity.GetLecturerProgramsResp, error) {
	type daoLecturer struct {
	}

	type daoTemplate struct {
	}

	var (
		resp = new(entity.GetLecturerProgramsResp)
	)
	resp.Items = make([]entity.LecturerProgramItem, 0)

	return resp, nil
}
