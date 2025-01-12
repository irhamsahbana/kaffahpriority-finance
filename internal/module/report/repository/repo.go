package repository

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/module/report/entity"
	"codebase-app/internal/module/report/ports"
	"context"

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
