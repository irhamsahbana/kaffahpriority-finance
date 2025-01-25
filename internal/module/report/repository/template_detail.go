package repository

import (
	"codebase-app/internal/module/report/entity"
	"codebase-app/pkg/errmsg"
	"context"
	"database/sql"

	"github.com/rs/zerolog/log"
)

func (r *reportRepo) GetTemplate(ctx context.Context, req *entity.GetTemplateReq) (*entity.GetTemplateResp, error) {
	var (
		resp = new(entity.GetTemplateResp)
	)
	resp.Students = make([]entity.AddStudent, 0)

	query := `
		SELECT
			prt.id,
			prt.user_id,
			prt.program_id,
			prt.lecturer_id,
			prt.marketer_id,
			prt.student_id,
			prt.days,
			prt.program_fee,
			prt.administration_fee,
			prt.foreign_learning_fee,
			prt.night_learning_fee,
			prt.marketer_commission_fee,
			prt.overpayment_fee,
			prt.hr_fee,
			prt.marketer_gifts_fee,
			prt.closing_fee_for_office,
			prt.closing_fee_for_reward,
			prt.notes,
			prt.created_at,
			prt.updated_at,
			prt.deleted_at,

			m.student_manager_id,
			l.name AS lecturer_name,
			m.name AS marketer_name,
			sm.name AS student_manager_name,
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
			student_managers sm
			ON m.student_manager_id = sm.id
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
