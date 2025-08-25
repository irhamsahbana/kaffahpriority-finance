package repository

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/module/report/entity"
	"codebase-app/internal/module/report/ports"
	"codebase-app/pkg/errmsg"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
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

func (r *reportRepo) GetLecturerPrograms(ctx context.Context, req *entity.GetLecturerProgramsReq) (*entity.GetLecturerProgramsResp, error) {
	type daoLecturer struct {
		TotalData int `db:"total_data"`
		entity.LecturerProgramItem
	}

	var (
		fnName       = "repo::GetLecturerPrograms"
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
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to fetch data", fnName)
		return nil, err
	}

	for _, item := range data {
		resp.Meta.TotalData = item.TotalData
		item.Templates = make([]entity.LecturerTemplate, 0)
		lecturerIds = append(lecturerIds, item.LecturerID)
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
			log.Error().Err(err).Any("req", req).Msgf("%s - failed to build query", fnName)
			return nil, err
		}

		query = r.db.Rebind(query)
		err = r.db.SelectContext(ctx, &dataTemplate, query, args...)
		if err != nil {
			log.Error().Err(err).Any("req", req).Msgf("%s - failed to fetch data", fnName)
			return nil, err
		}

		for _, item := range dataTemplate {
			mapTemplate[item.LecturerID] = append(mapTemplate[item.LecturerID], item)
		}

		for i, item := range resp.Items {
			resp.Items[i].Templates = mapTemplate[item.LecturerID]
		}
	}

	resp.Meta.CountTotalPage(req.Page, req.Paginate, resp.Meta.TotalData)

	return resp, nil
}

func (r *reportRepo) DistributeHRFee(ctx context.Context, req *entity.HRDistributionReq) error {
	fnName := "repo::DistributeHRFee"

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to start transaction", fnName)
		return err
	}
	defer tx.Rollback()

	// get HR fee
	queryFee := `SELECT hr_fee from program_registrations WHERE id = ? AND deleted_at IS NULL`
	var hrFee decimal.Decimal
	err = tx.GetContext(ctx, &hrFee, r.db.Rebind(queryFee), req.RegistrationID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warn().Any("req", req).Msgf("%s - HR fee not found", fnName)
			return errmsg.NewCustomErrors(404).SetMessage("Laporan tidak ditemukan")
		}
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to fetch HR fee", fnName)
		return err
	}

	hrFeeForMentor := decimal.NewFromFloat(req.HRFeeForMentor)
	hrFeeForHR := decimal.NewFromFloat(req.HRFeeForHR)

	if hrFeeForMentor.Add(hrFeeForHR).GreaterThan(hrFee) || hrFeeForMentor.Add(hrFeeForHR).LessThan(hrFee) {
		log.Warn().Any("req", req).Msgf("%s - HR fee is greater than total HR fee", fnName)
		return errmsg.NewCustomErrors(400).SetMessage("Distribusi Pengeluaran SDM tidak boleh melebihi atau kurang dari total biaya SDM")
	}

	query := `
		UPDATE
			program_registrations
		SET
			mentor_detail_fee = ?,
			hr_detail_fee = ?
		WHERE
			id = ?
			AND deleted_at IS NULL
		`

	_, err = tx.ExecContext(ctx, r.db.Rebind(query), req.HRFeeForMentor, req.HRFeeForHR, req.RegistrationID)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to update data", fnName)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to commit transaction", fnName)
		return err
	}

	return nil
}
