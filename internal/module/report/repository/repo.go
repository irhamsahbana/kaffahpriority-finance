package repository

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/module/report/entity"
	"codebase-app/internal/module/report/ports"
	"codebase-app/pkg/errmsg"
	"context"

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
		log.Error().Err(err).Any("req", req).Msg("repo::CreateTemplate - failed to check combination")
		return nil, err
	}

	if isCombinationExist {
		log.Warn().Any("req", req).Msg("repo::CreateTemplate - combination already exist")
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
			administration_fee,
			foreign_learning_fee,
			night_learning_fee,
			marketer_commission_fee,
			overpayment_fee,
			hr_fee,
			marketer_gifts_fee,
			closing_fee_for_office,
			closing_fee_for_reward
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?,
			(SELECT program_fee FROM program),
			?,
			?,
			?,
			(SELECT marketer_commission_fee FROM program),
			?,
			(SELECT hr_fee FROM program),
			?, ?, ?
		)
	`

	_, err = tx.ExecContext(ctx, tx.Rebind(query),
		req.ProgramId,
		Id, req.UserId, req.ProgramId, req.LecturerId, req.MarketerId, req.StudentId,
		pq.Array(req.Days), req.Notes,

		req.AdministrationFee,
		req.FLFee,
		req.NLFee,
		req.OverpaymentFee,
		req.MarketerGiftsFee,
		req.ClosingFeeForOffice,
		req.ClosingFeeForReward,
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
