package repository

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/module/report/entity"
	"codebase-app/internal/module/report/ports"
	"codebase-app/pkg/errmsg"
	"context"
	"database/sql"

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
		log.Error().Err(err).Any("req", req).Msg("repo::CreateTemplate - failed to begin transaction")
		return nil, err
	}
	defer func() {
		if err != nil {
			errRB := tx.Rollback()
			if errRB != nil {
				log.Error().Err(errRB).Any("req", req).Msg("repo::CreateTemplate - failed to rollback transaction")
			}
			return
		}
		errCommit := tx.Commit()
		if errCommit != nil {
			log.Error().Err(errCommit).Any("req", req).Msg("repo::CreateTemplate - failed to commit transaction")
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
				AND prt.lecturer_id = ?
				AND prt.marketer_id = ?
				AND prt.student_id = ?
				AND prt.deleted_at IS NULL
		)
	`

	err = tx.GetContext(ctx, &isCombinationExist, tx.Rebind(queryCheckCombination), req.ProgramId, req.LecturerId, req.MarketerId, req.StudentId)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::CreateTemplate - failed to check combination")
		return nil, err
	}

	if isCombinationExist {
		log.Warn().Any("req", req).Msg("repo::CreateTemplate - combination already exist")
		return nil, errmsg.NewCustomErrors(409).SetMessage("Template sudah ada")
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
				p.lecturer_fee AS hr_fee,
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
			hr_fee,
			marketer_commission_fee
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?,
			(SELECT program_fee FROM program),
			(SELECT hr_fee FROM program),
			(SELECT marketer_commission_fee FROM program)
		)
	`

	_, err = tx.ExecContext(ctx, tx.Rebind(query),
		req.ProgramId,
		Id, req.UserId, req.ProgramId, req.LecturerId, req.MarketerId, req.StudentId,
		pq.Array(req.Days), req.Notes,
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

func (r *reportRepo) GetRegistrations(ctx context.Context, req *entity.GetRegistrationsReq) (*entity.GetRegistrationsResp, error) {
	type dao struct {
		TotalData int `db:"total_data"`
		entity.RegisItem
	}
	var (
		data            = make([]dao, 0, req.Paginate)
		registrationIds = make([]string, 0)
		resp            = new(entity.GetRegistrationsResp)
		args            = make([]any, 0, 3)
	)
	resp.Items = make([]entity.RegisItem, 0)

	query := `
		SELECT
			COUNT(*) OVER() AS total_data,
			pr.id,
			pr.program_id,
			pr.marketer_id,
			pr.lecturer_id,
			pr.student_id,
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
			pr.paid_at,
			pr.created_at,
			pr.updated_at,
			pr.notes,
			pr.program_fee +
			COALESCE(pr.foreign_learning_fee, 0) +
			COALESCE(pr.night_learning_fee, 0) +
			COALESCE(pr.overpayment_fee, 0)
			AS monthly_fee,
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
			s.name AS student_name
		FROM
			program_registrations pr
		JOIN
			lecturers l
			ON pr.lecturer_id = l.id
		JOIN
			marketers m
			ON pr.marketer_id = m.id
		JOIN
			students s
			ON pr.student_id = s.id
		JOIN
			programs p
			ON pr.program_id = p.id
		WHERE
			pr.deleted_at IS NULL
	`

	if req.PaidAtFrom != "" && req.PaidAtTo != "" {
		query += `
			AND pr.paid_at AT TIME ZONE ? BETWEEN
			(TO_TIMESTAMP(?, 'YYYY-MM-DD') AT TIME ZONE 'UTC') AND
			(TO_TIMESTAMP(?, 'YYYY-MM-DD') AT TIME ZONE 'UTC' + time '23:59:59.999999')
		`
		args = append(args, req.Timezone, req.PaidAtFrom, req.PaidAtTo)
	}

	if req.Q != "" {
		query += ` AND (
			pr.program_name ILIKE '%' || ? || '%' OR
			s.name ILIKE '%' || ? || '%'
		)`
		args = append(args, req.Q, req.Q)
	}

	if req.LecturerId != "" {
		query += ` AND pr.lecturer_id = ?`
		args = append(args, req.LecturerId)
	}

	if req.MarketerId != "" {
		query += ` AND pr.marketer_id = ?`
		args = append(args, req.MarketerId)
	}

	if req.StudentId != "" {
		query += ` AND pr.student_id = ?`
		args = append(args, req.StudentId)
	}

	if req.ProgramId != "" {
		query += ` AND pr.program_id = ?`
		args = append(args, req.ProgramId)
	}

	sortByMap := map[string]string{
		"created_at": "pr.created_at",
		"paid_at":    "pr.paid_at",
		"updated_at": "pr.updated_at",
		"":           "pr.paid_at",
	}

	sortTypeMap := map[string]string{
		"asc":  "ASC",
		"desc": "DESC",
		"":     "DESC",
	}

	query += ` ORDER BY ` + sortByMap[req.SortBy] + ` ` + sortTypeMap[req.SortType] + ` LIMIT ? OFFSET ?`
	args = append(args, req.Paginate, (req.Page-1)*req.Paginate)

	err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), args...)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetRegistrations - failed to fetch data")
		return nil, err
	}

	for _, item := range data {
		resp.Meta.TotalData = item.TotalData
		registrationIds = append(registrationIds, item.Id)
		resp.Items = append(resp.Items, item.RegisItem)
		resp.Items[len(resp.Items)-1].Students = make([]entity.AddStudent, 0)
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	if len(registrationIds) > 0 {
		type daos struct {
			PrId string `db:"pr_id"`
			entity.AddStudent
		}
		var (
			daosData = make([]daos, 0)
		)
		query = `
			SELECT
				prs.pr_id,
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
			WHERE prs.pr_id IN (?)
		`

		query, args, err := sqlx.In(query, registrationIds)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msg("repo::GetRegistrations - failed to build query")
			return nil, err
		}

		query = r.db.Rebind(query)
		err = r.db.SelectContext(ctx, &daosData, query, args...)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msg("repo::GetRegistrations - failed to fetch additional students")
			return nil, err
		}

		for i, item := range resp.Items {
			for _, data := range daosData {
				if item.Id == data.PrId {
					resp.Items[i].Students = append(resp.Items[i].Students, data.AddStudent)
				}
			}
		}
	}

	return resp, nil
}

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
		JOIN
			lecturers l
			ON pr.lecturer_id = l.id
		JOIN
			marketers m
			ON pr.marketer_id = m.id
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

func (r *reportRepo) GetLecturerPrograms(ctx context.Context, req *entity.GetLecturerProgramsReq) (*entity.GetLecturerProgramsResp, error) {
	type daoLecturer struct {
		TotalData int `db:"total_data"`
		entity.LecturerProgramItem
	}

	var (
		resp         = new(entity.GetLecturerProgramsResp)
		data         = make([]daoLecturer, 0, req.Paginate)
		dataTemplate = make([]entity.LecturerTemplate, req.Paginate)
		lecturerIds  = make([]string, 0)
		mapTemplate  = make(map[string][]entity.LecturerTemplate)
	)
	resp.Items = make([]entity.LecturerProgramItem, 0, req.Paginate)

	query := `
		SELECT
			COUNT(*) OVER() AS total_data,
			prt.lecturer_id AS lecturer_id,
			l.name AS lecturer_name
		FROM
			program_registration_templates prt
		JOIN
			lecturers l
			ON prt.lecturer_id = l.id
		WHERE
			prt.deleted_at IS NULL
		`

	if req.IsFinanceUpdated != "" {
		if req.IsFinanceUpdated == "true" {
			query += ` AND prt.program_fee IS NOT NULL`
		} else {
			query += ` AND prt.program_fee IS NULL`
		}
	}

	query += `
		GROUP BY
			prt.lecturer_id, l.name
		ORDER BY
			l.name
		LIMIT ? OFFSET ?
	`

	err := r.db.SelectContext(ctx, &data, r.db.Rebind(query), req.Paginate, (req.Page-1)*req.Paginate)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetLecturerPrograms - failed to fetch data")
		return nil, err
	}

	for _, item := range data {
		resp.Meta.TotalData = item.TotalData
		item.Templates = make([]entity.LecturerTemplate, 0)
		lecturerIds = append(lecturerIds, item.LecturerId)
		resp.Items = append(resp.Items, item.LecturerProgramItem)
	}

	if len(resp.Items) > 0 {
		query = `
			SELECT
				prt.lecturer_id,
				prt.id AS template_id,
				prt.program_id,
				prt.student_id,
				prt.marketer_id,
				p.name AS program_name,
				s.name AS student_name,
				m.name AS marketer_name,
				COALESCE(prt.program_fee, 0) +
				COALESCE(prt.foreign_learning_fee, 0) +
				COALESCE(prt.night_learning_fee, 0) +
				COALESCE(prt.overpayment_fee, 0)
				AS monthly_fee,
				CASE
					WHEN prt.program_fee IS NULL THEN FALSE
					ELSE TRUE
				END AS is_finance_updated
			FROM
				program_registration_templates prt
			JOIN
				programs p
				ON prt.program_id = p.id
			JOIN
				students s
				ON prt.student_id = s.id
			JOIN
				marketers m
				ON prt.marketer_id = m.id
			WHERE
				prt.lecturer_id IN (?)
			`

		query, args, err := sqlx.In(query, lecturerIds)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msg("repo::GetLecturerPrograms - failed to build query")
			return nil, err
		}

		query = r.db.Rebind(query)
		err = r.db.SelectContext(ctx, &dataTemplate, query, args...)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msg("repo::GetLecturerPrograms - failed to fetch data")
			return nil, err
		}

		for _, item := range dataTemplate {
			mapTemplate[item.LecturerId] = append(mapTemplate[item.LecturerId], item)
		}

		for i, item := range resp.Items {
			resp.Items[i].Templates = mapTemplate[item.LecturerId]
		}
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	return resp, nil
}
