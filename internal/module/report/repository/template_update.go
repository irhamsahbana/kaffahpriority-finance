package repository

import (
	"codebase-app/internal/module/report/entity"
	"codebase-app/pkg/errmsg"
	"context"

	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (r *reportRepo) UpdateTemplate(ctx context.Context, req *entity.UpdateTemplateGeneralReq) (*entity.UpdateTemplateResp, error) {
	fnName := "repo::UpdateTemplate"
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

	var isCombinationExist bool
	queryCombination := `
		SELECT EXISTS (
			SELECT 1
			FROM program_registration_templates prt
			WHERE
				prt.program_id = ?
				AND prt.marketer_id = ?
				AND prt.student_id = ?
				AND (
					(prt.lecturer_id IS NULL AND ?::TEXT IS NULL)
					OR prt.lecturer_id = ?
				)
				AND prt.id != ?
				AND prt.deleted_at IS NULL
		)
	`

	err = tx.GetContext(ctx, &isCombinationExist, tx.Rebind(queryCombination),
		req.ProgramId, req.MarketerId, req.StudentId, req.LecturerId, req.LecturerId, req.Id,
	)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to check combination", fnName)
		return nil, err
	}

	if isCombinationExist {
		log.Warn().Any("req", req).Msgf("%s - combination already exist", fnName)
		return nil, errmsg.NewCustomErrors(409).SetMessage("Template dengan kombinasi program, marketer, pengajar, dan santri tersebut sudah ada. Silahkan cek kembali atau update data yang sudah ada")
	}

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
		req.Id,
	)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to update data", fnName)
		return nil, err
	}

	query = `
		DELETE FROM prt_additional_students WHERE prt_id = ?
	`
	_, err = tx.ExecContext(ctx, tx.Rebind(query), req.Id)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to delete additional students", fnName)
		return nil, err
	}
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to delete additional students", fnName)
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
			log.Error().Err(err).Any("req", req).Msgf("%s - failed to insert additional students", fnName)
			return nil, err
		}
	}

	resp := new(entity.UpdateTemplateResp)
	resp.Id = req.Id

	return resp, nil
}
