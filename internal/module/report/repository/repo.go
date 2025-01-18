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
		log.Error().Err(err).Msg("repo::CreateTemplate - failed to begin transaction")
		return nil, err
	}
	defer func() {
		if err != nil {
			errRB := tx.Rollback()
			if errRB != nil {
				log.Error().Err(errRB).Msg("repo::CreateTemplate - failed to rollback transaction")
			}
			return
		}
		errCommit := tx.Commit()
		if errCommit != nil {
			log.Error().Err(errCommit).Msg("repo::CreateTemplate - failed to commit transaction")
		}
	}()

	var (
		Id   = ulid.Make().String()
		resp = new(entity.CreateTemplateResp)
	)
	resp.Id = Id

	query := `
		INSERT INTO program_registration_templates (
			id, user_id, program_id, lecturer_id, marketer_id, student_id,
			program_fee,
			administration_fee,
			foreign_lecturer_fee,
			night_learning_fee,
			marketer_commission_fee,
			overpayment_fee,
			hr_fee,
			marketer_gifts_fee,
			closing_fee_for_office,
			closing_fee_for_reward,
			days
		) VALUES (
		 	?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	_, err = tx.ExecContext(ctx, tx.Rebind(query),
		Id, req.UserId, req.ProgramId, req.LecturerId, req.MarketerId, req.StudentId,
		req.ProgramFee, req.AdministrationFee, req.FLFee, req.NLFee,
		req.MarketerCommissionFee, req.OverpaymentFee, req.HRFee, req.MarketerGiftsFee,
		req.ClosingFeeForOffice, req.ClosingFeeForReward,
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

func (r *reportRepo) UpdateTemplate(ctx context.Context, req *entity.UpdateTemplateReq) (*entity.UpdateTemplateResp, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("repo::UpdateTemplate - failed to begin transaction")
		return nil, err
	}
	defer func() {
		if err != nil {
			errRB := tx.Rollback()
			if errRB != nil {
				log.Error().Err(errRB).Msg("repo::UpdateTemplate - failed to rollback transaction")
			}
			return
		}
		errCommit := tx.Commit()
		if errCommit != nil {
			log.Error().Err(errCommit).Msg("repo::UpdateTemplate - failed to commit transaction")
		}
	}()

	query := `
		UPDATE program_registration_templates SET
			program_id = ?,
			lecturer_id = ?,
			marketer_id = ?,
			student_id = ?,
			program_fee = ?,
			administration_fee = ?,
			foreign_lecturer_fee = ?,
			night_learning_fee = ?,
			marketer_commission_fee = ?,
			overpayment_fee = ?,
			hr_fee = ?,
			marketer_gifts_fee = ?,
			closing_fee_for_office = ?,
			closing_fee_for_reward = ?,
			days = ?,
			updated_at = NOW()
		WHERE
			id = ?
			AND deleted_at IS NULL
	`

	_, err = tx.ExecContext(ctx, tx.Rebind(query),
		req.ProgramId, req.LecturerId, req.MarketerId, req.StudentId,
		req.ProgramFee, req.AdministrationFee, req.FLFee, req.NLFee,
		req.MarketerCommissionFee, req.OverpaymentFee, req.HRFee, req.MarketerGiftsFee,
		req.ClosingFeeForOffice, req.ClosingFeeForReward,
		pq.Array(req.Days),
		req.Id,
	)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateTemplate - failed to update data")
		return nil, err
	}

	query = `
		DELETE FROM prt_additional_students WHERE prt_id = ?
	`
	_, err = tx.ExecContext(ctx, tx.Rebind(query), req.Id)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateTemplate - failed to delete additional students")
		return nil, err
	}
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UpdateTemplate - failed to delete additional students")
		return nil, err
	}

	for _, item := range req.AdditionalStudents {
		query = `
			INSERT INTO prt_additional_students (
				id, prt_id, student_id, name
			) VALUES (?, ?, ?, ?)
		`

		_, err = tx.ExecContext(ctx, tx.Rebind(query),
			ulid.Make().String(), req.Id, item.StudentId, item.Name,
		)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msg("repo::UpdateTemplate - failed to insert additional students")
			return nil, err
		}
	}

	resp := new(entity.UpdateTemplateResp)
	resp.Id = req.Id

	return resp, nil
}

func (r *reportRepo) GetTemplates(ctx context.Context, req *entity.GetTemplatesReq) (*entity.GetTemplatesResp, error) {
	type dao struct {
		TotalData int `db:"total_data"`
		entity.TemplateItem
	}
	var (
		data        = make([]dao, 0, req.Paginate)
		templateIds = make([]string, 0)
		resp        = new(entity.GetTemplatesResp)
	)
	resp.Items = make([]entity.TemplateItem, 0)

	query := `
		SELECT
			COUNT(*) OVER() AS total_data,
			/*
			prt.id,
			prt.marketer_id,
			prt.lecturer_id,
			prt.student_id,
			prt.created_at,
			prt.updated_at,
			*/
			prt.*,
			p.name AS program_name,
			l.name AS lecturer_name,
			m.name AS marketer_name,
			s.name AS student_name,
			prt.program_fee +
			COALESCE(prt.foreign_lecturer_fee, 0) +
			COALESCE(prt.night_learning_fee, 0) +
			COALESCE(prt.overpayment_fee, 0)
			AS monthly_fee
		FROM
			program_registration_templates prt
		JOIN
			lecturers l
			ON prt.lecturer_id = l.id
		JOIN
			marketers m
			ON prt.marketer_id = m.id
		JOIN
			students s
			ON prt.student_id = s.id
		JOIN
			programs p
			ON prt.program_id = p.id
		WHERE
			prt.deleted_at IS NULL
	`

	err := r.db.SelectContext(ctx, &data, r.db.Rebind(query))
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetTemplates - failed to fetch data")
		return nil, err
	}

	for _, item := range data {
		resp.Meta.TotalData = item.TotalData
		templateIds = append(templateIds, item.Id)
		item.Students = make([]entity.AddStudent, 0)
		resp.Items = append(resp.Items, item.TemplateItem)
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	if len(templateIds) > 0 {
		type daos struct {
			PrtId string `db:"prt_id"`
			entity.AddStudent
		}
		var (
			daosData = make([]daos, 0)
		)
		query = `
			SELECT
				adds.prt_id,
				adds.student_id,
				CASE
					WHEN s.id IS NULL THEN adds.name
					ELSE s.name
				END AS name
			FROM
				prt_additional_students adds
			LEFT JOIN
				students s
				ON adds.student_id = s.id
			WHERE adds.prt_id IN (?)
		`

		query, args, err := sqlx.In(query, templateIds)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msg("repo::GetTemplates - failed to build query")
			return nil, err
		}

		query = r.db.Rebind(query)
		err = r.db.SelectContext(ctx, &daosData, query, args...)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msg("repo::GetTemplates - failed to fetch additional students")
			return nil, err
		}

		for i, item := range resp.Items {
			for _, data := range daosData {
				if item.Id == data.PrtId {
					resp.Items[i].Students = append(resp.Items[i].Students, data.AddStudent)
				}
			}
		}
	}

	return resp, nil
}

func (r *reportRepo) GetTemplate(ctx context.Context, req *entity.GetTemplateReq) (*entity.GetTemplateResp, error) {
	var (
		resp = new(entity.GetTemplateResp)
	)
	resp.Students = make([]entity.AddStudent, 0)

	query := `
		SELECT
			prt.*,
			l.name AS lecturer_name,
			m.name AS marketer_name,
			s.name AS student_name,
			p.name AS program_name
		FROM
			program_registration_templates prt
		JOIN
			lecturers l
			ON prt.lecturer_id = l.id
		JOIN
			marketers m
			ON prt.marketer_id = m.id
		JOIN
			students s
			ON prt.student_id = s.id
		JOIN
			programs p
			ON prt.program_id = p.id
		WHERE
			prt.id = ?
			AND prt.deleted_at IS NULL
	`

	err := r.db.GetContext(ctx, resp, r.db.Rebind(query), req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warn().Err(err).Any("req", req).Msg("repo::GetTemplate - data not found")
			return nil, errmsg.NewCustomErrors(404).SetMessage("Template tidak ditemukan")
		}
		log.Error().Err(err).Any("req", req).Msg("repo::GetTemplate - failed to fetch data")
		return nil, err
	}

	query = `
		SELECT
			adds.student_id,
			CASE
				WHEN s.id IS NULL THEN adds.name
				ELSE s.name
			END AS name
		FROM
			prt_additional_students adds
		LEFT JOIN
			students s
			ON adds.student_id = s.id
		WHERE
			adds.prt_id = ?
	`

	err = r.db.SelectContext(ctx, &resp.Students, r.db.Rebind(query), req.Id)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetTemplate - failed to fetch additional students")
		return nil, err
	}

	return resp, nil
}
