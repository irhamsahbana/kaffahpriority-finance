package repository

import (
	"codebase-app/internal/module/report/entity"
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

func (r *reportRepo) GetTemplates(ctx context.Context, req *entity.GetTemplatesReq) (*entity.GetTemplatesResp, error) {
	type dao struct {
		TotalData int `db:"total_data"`
		entity.TemplateItem
	}
	var (
		data        = make([]dao, 0, req.Paginate)
		templateIds = make([]string, 0)
		resp        = new(entity.GetTemplatesResp)
		args        = make([]any, 0, 2)
	)
	resp.Items = make([]entity.TemplateItem, 0)

	query := `
		SELECT
			COUNT(*) OVER() AS total_data,
			prt.id,
			prt.user_id,
			prt.program_id,
			prt.lecturer_id,
			prt.marketer_id,
			prt.student_id,
			s.identifier as student_identifier,
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
			p.name AS program_name,
			l.name AS lecturer_name,
			m.name AS marketer_name,
			s.name AS student_name,
			sm.name AS student_manager_name,
			COALESCE(prt.program_fee, 0) +
			COALESCE(prt.foreign_learning_fee, 0) +
			COALESCE(prt.night_learning_fee, 0) +
			COALESCE(prt.overpayment_fee, 0)
			AS monthly_fee,
			CASE
				WHEN prt.is_financially_cleared = FALSE THEN TRUE
				ELSE FALSE
			END AS is_finance_update_required
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
			prt.deleted_at IS NULL
	`

	if req.IsFinanceUpdateRequired != "" {
		if req.IsFinanceUpdateRequired == "true" {
			query += ` AND prt.is_financially_cleared = FALSE `
		} else {
			query += ` AND prt.is_financially_cleared = TRUE `
		}
	}

	if req.MarketerId != "" {
		query += ` AND prt.marketer_id = ? `
		args = append(args, req.MarketerId)
	}

	if req.StudentManagerId != "" {
		query += ` AND m.student_manager_id = ? `
		args = append(args, req.StudentManagerId)
	}

	if req.LecturerId != "" {
		query += ` AND prt.lecturer_id = ? `
		args = append(args, req.LecturerId)
	}

	if req.StudentId != "" {
		query += ` AND prt.student_id = ? `
		args = append(args, req.StudentId)
	}

	if req.ProgramId != "" {
		query += ` AND prt.program_id = ? `
		args = append(args, req.ProgramId)
	}

	sortMap := map[string]string{
		"asc":  "ASC",
		"desc": "DESC",
		"":     "DESC",
	}

	sortByMap := map[string]string{
		"created_at": "prt.created_at",
		"updated_at": "prt.updated_at",
		"":           "prt.updated_at",
	}

	query += ` ORDER BY ` + sortByMap[req.SortBy] + ` ` + sortMap[req.SortType] + ` LIMIT ? OFFSET ? `
	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...)
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
