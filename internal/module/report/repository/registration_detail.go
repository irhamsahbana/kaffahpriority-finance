package repository

import (
	"codebase-app/internal/module/report/entity"
	"codebase-app/pkg/errmsg"
	"context"
	"database/sql"

	"github.com/rs/zerolog/log"
)

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
			m.student_manager_id,
			s.identifier AS student_identifier,
			sm.name AS student_manager_name,
			pr.days,
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
			pr.notes,
			(
				COALESCE(pr.administration_fee, 0)
				+ COALESCE(pr.program_fee, 0)
				+ COALESCE(pr.overpayment_fee, 0)
				+ COALESCE(pr.night_learning_fee, 0)
				+ COALESCE(pr.foreign_learning_fee, 0)
				- COALESCE(pr.marketer_commission_fee, 0)
				- COALESCE(pr.marketer_gifts_fee, 0)
				- COALESCE(pr.hr_fee, 0)
				- COALESCE(pr.overpayment_fee, 0)
				- COALESCE(pr.closing_fee_for_office, 0)
				- COALESCE(pr.closing_fee_for_reward, 0)
			) AS profit,
			l.name AS lecturer_name,
			m.name AS marketer_name,
			s.name AS student_name,
			p.name AS program_name
		FROM
			program_registrations pr
		LEFT JOIN
			lecturers l
			ON pr.lecturer_id = l.id
		JOIN
			marketers m
			ON pr.marketer_id = m.id
		JOIN
			student_managers sm
			ON m.student_manager_id = sm.id
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
