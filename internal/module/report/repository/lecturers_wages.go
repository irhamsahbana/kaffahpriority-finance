package repository

import (
	"codebase-app/internal/module/report/entity"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *reportRepo) GetLecturersWages(ctx context.Context, req *entity.GetLecturersWagesReq) (*entity.GetLecturersWagesResp, error) {
	type dao struct {
		TotalData int `db:"total_data"`
		entity.LecturersWageItem
	}
	var (
		resp = new(entity.GetLecturersWagesResp)
		data = make([]dao, 0)
		args = make([]any, 0, 3)
	)

	resp.Items = make([]entity.LecturersWageItem, 0)

	query := `
		SELECT
			COUNT (*) OVER() AS total_data,
			pr.id AS registration_id,
			pr.program_name,
			s.name AS student_name,
			l.name AS lecturer_name,
			pr.foreign_learning_fee,
			pr.night_learning_fee,
			pr.is_itp,
			pr.program_meetings,
			pr.program_fee_per_meeting,
			pr.is_full_fee,
			pr.full_fee,
			pr.mentor_detail_fee_used
		FROM
			program_registrations pr
		JOIN
			programs p ON pr.program_id = p.id
		JOIN
			lecturers l ON pr.lecturer_id = l.id
		JOIN
			students s ON pr.student_id = s.id
		WHERE
			pr.deleted_at IS NULL
			AND TO_CHAR(pr.paid_at AT TIME ZONE ?, 'YYYY-MM') = ?
	`

	args = append(args, req.Timezone, req.Month)

	if req.LecturerId != "" {
		query += ` AND pr.lecturer_id = ?`
		args = append(args, req.LecturerId)
	}

	query += `
		ORDER BY
			pr.lecturer_id ASC,
			pr.student_id ASC,
			pr.program_name ASC
		LIMIT ? OFFSET ?
	`
	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	if err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetLecturersWages - failed to query lecturers wages")
		return nil, err
	}

	for _, d := range data {
		resp.Meta.TotalData = d.TotalData
		resp.Items = append(resp.Items, d.LecturersWageItem)
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)
	return resp, nil
}
