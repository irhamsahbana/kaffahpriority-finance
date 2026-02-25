package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"

	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (r *reportRepo) UpdateTemplate(ctx context.Context, req *entity.UpdateTemplateGeneralReq) (*entity.UpdateTemplateResp, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.UpdateTemplate")
	defer span.End()
	fnName := "repo::UpdateTemplate"
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("%s - failed to begin transaction", fnName)
		return nil, err
	}
	defer func() {
		if err != nil {
			errRB := tx.Rollback()
			if errRB != nil {
				log.Ctx(ctx).Error().Err(errRB).Msgf("%s - failed to rollback transaction", fnName)
			}
			return
		}
		errCommit := tx.Commit()
		if errCommit != nil {
			log.Ctx(ctx).Error().Err(errCommit).Msgf("%s - failed to commit transaction", fnName)
		}
	}()

	query := `
		UPDATE program_registration_templates SET
			program_id = ?,
			lecturer_id = ?,
			marketer_id = ?,
			student_id = ?,
			days = ?,
			notes = ?,

			program_fee = ?,
			administration_fee = ?,
			foreign_learning_fee = ?,
			night_learning_fee = ?,
			is_itp = ?,
			marketer_commission_fee = ?,
			overpayment_fee = ?,
			hr_fee = ?,
			marketer_gifts_fee = ?,
			closing_fee_for_office = ?,
			closing_fee_for_reward = ?,
			created_at = COALESCE(?, created_at),
			updated_at = NOW()
		WHERE
			id = ?
			AND deleted_at IS NULL
	`

	_, err = tx.ExecContext(ctx, tx.Rebind(query),
		req.ProgramId, req.LecturerId, req.MarketerId, req.StudentId,
		pq.Array(req.Days), req.Notes,
		req.ProgramFee, req.AdministrationFee, req.FLFee, req.NLFee, req.IsITP,
		req.MarketerCommissionFee, req.OverpaymentFee, req.HRFee, req.MarketerGiftsFee,
		req.ClosingFeeForOffice, req.ClosingFeeForReward,
		req.CreatedAt,
		req.ID,
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to update data", fnName)
		return nil, err
	}

	query = `
		DELETE FROM prt_additional_students WHERE prt_id = ?
	`
	_, err = tx.ExecContext(ctx, tx.Rebind(query), req.ID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to delete additional students", fnName)
		return nil, err
	}
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to delete additional students", fnName)
		return nil, err
	}

	for _, item := range req.AdditionalStudents {
		query = `
			INSERT INTO prt_additional_students (
				id, prt_id, student_id, name
			) VALUES (?, ?, ?, ?)
		`

		_, err = tx.ExecContext(ctx, tx.Rebind(query),
			ulid.Make().String(), req.ID, item.StudentID, item.Name,
		)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to insert additional students", fnName)
			return nil, err
		}
	}

	resp := new(entity.UpdateTemplateResp)
	resp.ID = req.ID

	return resp, nil
}
