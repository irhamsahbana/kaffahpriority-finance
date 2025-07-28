package repository

import (
	"codebase-app/internal/module/report/entity"
	"codebase-app/pkg/errmsg"
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (r *reportRepo) UpdateRegistration(ctx context.Context, req *entity.UpdateRegistrationReq) (*entity.UpdateRegistrationResp, error) {
	fnName := "repo::UpdateRegistration"
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msgf("%s - failed to begin transaction", fnName)
		return nil, err
	}
	defer func() {
		if err != nil {
			errRB := tx.Rollback()
			if errRB != nil {
				log.Error().Err(errRB).Msgf("%s - failed to rollback transaction", fnName)
			}
			return
		}
		errCommit := tx.Commit()
		if errCommit != nil {
			log.Error().Err(errCommit).Msgf("%s - failed to commit transaction", fnName)
		}
	}()

	query := `
		UPDATE program_registrations SET
			program_id = ?,
			lecturer_id = ?,
			marketer_id = ?,
			student_id = ?,
			program_name = (SELECT name FROM programs WHERE id = ?),
			program_fee = ?,
			administration_fee = ?,
			foreign_learning_fee = ?,
			night_learning_fee = ?,
			marketer_commission_fee = ?,
			overpayment_fee = ?,
			hr_fee = ?,
			marketer_gifts_fee = ?,
			closing_fee_for_office = ?,
			closing_fee_for_reward = ?,
			days = ?,
			notes = ?,
			is_itp = ?,
			updated_at = NOW()
		WHERE
			id = ?
			AND deleted_at IS NULL
	`

	_, err = tx.ExecContext(ctx, tx.Rebind(query),
		req.ProgramId, req.LecturerId, req.MarketerId, req.StudentId,
		req.ProgramId, req.ProgramFee, req.AdministrationFee, req.FLFee, req.NLFee,
		req.MarketerCommissionFee, req.OverpaymentFee, req.HRFee, req.MarketerGiftsFee,
		req.ClosingFeeForOffice, req.ClosingFeeForReward, pq.Array(req.Days), req.Notes,
		req.IsITP,
		req.Id,
	)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to update data", fnName)
		return nil, err
	}

	query = `
		DELETE FROM pr_additional_students WHERE pr_id = ?
	`
	_, err = tx.ExecContext(ctx, tx.Rebind(query), req.Id)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to delete additional students", fnName)
		return nil, err
	}

	for _, item := range req.Students {
		query = `
			INSERT INTO pr_additional_students (
				id, pr_id, student_id, name
			) VALUES (?, ?, ?, ?)
		`

		_, err = tx.ExecContext(ctx, tx.Rebind(query),
			ulid.Make().String(), req.Id, item.StudentId, item.Name,
		)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msgf("%s - failed to insert additional students", fnName)
			return nil, err
		}
	}

	if !req.IsUpdateTemplate {
		resp := new(entity.UpdateRegistrationResp)
		resp.Id = req.Id
		return resp, nil
	}

	// update template
	type registration struct {
		TemplateId string `db:"template_id"`
	}
	var reg registration
	query = `
		SELECT
			template_id
		FROM
			program_registrations
		WHERE
			id = ?
			AND deleted_at IS NULL
	`

	err = tx.GetContext(ctx, &reg, tx.Rebind(query), req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warn().Err(err).Any("req", req).Msgf("%s - registration not found", fnName)
			return nil, errmsg.NewCustomErrors(404).SetMessage("template tidak ditemukan")
		}
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to get registration data", fnName)
		return nil, err
	}

	query = `
		UPDATE program_registration_templates SET
			program_id = ?,
			lecturer_id = ?,
			marketer_id = ?,
			student_id = ?,
			program_fee = ?,
			administration_fee = ?,
			foreign_learning_fee = ?,
			night_learning_fee = ?,
			marketer_commission_fee = ?,
			overpayment_fee = ?,
			hr_fee = ?,
			marketer_gifts_fee = ?,
			closing_fee_for_office = ?,
			closing_fee_for_reward = ?,
			days = ?,
			notes = ?,
			is_itp = ?,
			updated_at = NOW()
		WHERE
			id = ?
			AND deleted_at IS NULL
	`

	_, err = tx.ExecContext(ctx, tx.Rebind(query),
		req.ProgramId, req.LecturerId, req.MarketerId, req.StudentId,
		req.ProgramFee, req.AdministrationFee, req.FLFee, req.NLFee,
		req.MarketerCommissionFee, req.OverpaymentFee, req.HRFee, req.MarketerGiftsFee,
		req.ClosingFeeForOffice, req.ClosingFeeForReward, pq.Array(req.Days), req.Notes,
		req.IsITP,
		reg.TemplateId,
	)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to update template data", fnName)
		return nil, err
	}

	query = `
		DELETE FROM prt_additional_students WHERE prt_id = ?
	`
	_, err = tx.ExecContext(ctx, tx.Rebind(query), reg.TemplateId)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to delete additional students from template", fnName)
		return nil, err
	}

	for _, item := range req.Students {
		query = `
			INSERT INTO prt_additional_students (
				id, prt_id, student_id, name
			) VALUES (?, ?, ?, ?)
		`

		_, err = tx.ExecContext(ctx, tx.Rebind(query),
			ulid.Make().String(), reg.TemplateId, item.StudentId, item.Name,
		)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msgf("%s - failed to insert additional students into template", fnName)
			return nil, err
		}
	}

	resp := new(entity.UpdateRegistrationResp)
	resp.Id = req.Id

	return resp, nil
}
