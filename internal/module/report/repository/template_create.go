package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"

	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (r *reportRepo) CreateTemplate(ctx context.Context, req *entity.CreateTemplateReq) (*entity.CreateTemplateResp, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.CreateTemplate")
	defer span.End()
	fnName := "repo::CreateTemplate"

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to begin transaction", fnName)
		return nil, err
	}
	defer func() {
		if err != nil {
			errRB := tx.Rollback()
			if errRB != nil {
				log.Ctx(ctx).Error().Err(errRB).Any("req", req).Msgf("%s - failed to rollback transaction", fnName)
			}
			return
		}
		errCommit := tx.Commit()
		if errCommit != nil {
			log.Ctx(ctx).Error().Err(errCommit).Any("req", req).Msgf("%s - failed to commit transaction", fnName)
		}
	}()

	var (
		Id   = ulid.Make().String()
		resp = new(entity.CreateTemplateResp)
	)
	resp.ID = Id

	query := `
		WITH program AS (
			SELECT
				p.price AS program_fee,
				p.price_per_meeting AS program_fee_per_meeting,
				p.commission_fee AS marketer_commission_fee
			FROM
				programs p
			WHERE
				p.id = ?
				AND p.deleted_at IS NULL
		)
		INSERT INTO program_registration_templates (
			id,
			user_id,
			program_id,
			lecturer_id,
			marketer_id,
			student_id,
			days,
			notes,
			program_fee,

			program_fee_per_meeting,
			administration_fee,
			foreign_learning_fee,
			night_learning_fee,
			is_itp,
			marketer_commission_fee,
			overpayment_fee,
			hr_fee,
			marketer_gifts_fee,
			closing_fee_for_office,
			closing_fee_for_reward
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?,
			(SELECT program_fee_per_meeting FROM program),
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			?, ?, ?
		)
	`

	_, err = tx.ExecContext(ctx, tx.Rebind(query),
		req.ProgramID,
		Id, req.UserID, req.ProgramID, req.LecturerID, req.MarketerID, req.StudentID,
		pq.Array(req.Days), req.Notes, req.ProgramFee,

		req.AdministrationFee,
		req.FLFee,
		req.NLFee,
		req.IsITP,
		req.MarketerCommissionFee,
		req.OverpaymentFee,
		req.HRFee,
		req.MarketerGiftsFee,
		req.ClosingFeeForOffice,
		req.ClosingFeeForReward,
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to insert data", fnName)
		return nil, err
	}

	for _, item := range req.AdditionalStudents {
		query = `
			INSERT INTO prt_additional_students (
				id, prt_id, student_id, name
			) VALUES (?, ?, ?, ?)
		`

		_, err = tx.ExecContext(ctx, tx.Rebind(query),
			ulid.Make().String(), Id, item.StudentID, item.Name,
		)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to insert additional students", fnName)
			return nil, err
		}
	}

	return resp, nil
}
