package repository

import (
	"codebase-app/internal/module/report/entity"
	"codebase-app/pkg/errmsg"
	"context"

	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (r *reportRepo) CreateTemplate(ctx context.Context, req *entity.CreateTemplateReq) (*entity.CreateTemplateResp, error) {
	fnName := "repo::CreateTemplate"

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to begin transaction", fnName)
		return nil, err
	}
	defer func() {
		if err != nil {
			errRB := tx.Rollback()
			if errRB != nil {
				log.Error().Err(errRB).Any("req", req).Msgf("%s - failed to rollback transaction", fnName)
			}
			return
		}
		errCommit := tx.Commit()
		if errCommit != nil {
			log.Error().Err(errCommit).Any("req", req).Msgf("%s - failed to commit transaction", fnName)
		}
	}()

	isCombinationExist := false

	queryCheckCombination := `
		SELECT EXISTS (
			SELECT
				1
			FROM
				program_registration_templates prt
			WHERE
				prt.program_id = ?
				AND prt.marketer_id = ?
				AND prt.student_id = ?
				AND (
					(prt.lecturer_id IS NULL AND ?::TEXT IS NULL)
					OR prt.lecturer_id = ?
				)
				AND prt.deleted_at IS NULL
		)
	`

	err = tx.GetContext(ctx, &isCombinationExist, tx.Rebind(queryCheckCombination),
		req.ProgramId, req.MarketerId, req.StudentId, req.LecturerId, req.LecturerId,
	)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to check combination", fnName)
		return nil, err
	}

	if isCombinationExist {
		log.Warn().Any("req", req).Msgf("%s - combination already exist", fnName)
		return nil, errmsg.NewCustomErrors(409).SetMessage("Template dengan kombinasi program, marketer, pengajar, dan santri tersebut sudah ada. Silahkan cek kembali atau update data yang sudah ada")
	}

	var (
		Id   = ulid.Make().String()
		resp = new(entity.CreateTemplateResp)
	)
	resp.Id = Id

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
			(SELECT marketer_commission_fee FROM program),
			?,
			?,
			?, ?, ?
		)
	`

	_, err = tx.ExecContext(ctx, tx.Rebind(query),
		req.ProgramId,
		Id, req.UserId, req.ProgramId, req.LecturerId, req.MarketerId, req.StudentId,
		pq.Array(req.Days), req.Notes, req.ProgramFee,

		req.AdministrationFee,
		req.FLFee,
		req.NLFee,
		req.IsITP,
		req.OverpaymentFee,
		req.HRFee,
		req.MarketerGiftsFee,
		req.ClosingFeeForOffice,
		req.ClosingFeeForReward,
	)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to insert data", fnName)
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
			log.Error().Err(err).Any("req", req).Msgf("%s - failed to insert additional students", fnName)
			return nil, err
		}
	}

	return resp, nil
}
