package repository

import (
	"codebase-app/internal/module/report/entity"
	"codebase-app/pkg/errmsg"
	"context"
	"database/sql"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

func (r *reportRepo) UseHRfeeForLecturer(ctx context.Context, req *entity.UseHRfeeForLecturerReq) error {
	Tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UseHRfeeForLLecturer - failed to begin transaction")
		return err
	}
	defer Tx.Rollback()

	queryGetMentorDetailFee := `
		SELECT
			mentor_detail_fee
		FROM
			program_registrations
		WHERE
			id = ?
	`
	var mentorDetailFee decimal.Decimal

	err = Tx.GetContext(ctx, &mentorDetailFee, Tx.Rebind(queryGetMentorDetailFee), req.RegistrationId)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warn().Err(err).Any("req", req).Msg("repo::UseHRfeeForLecturer - mentor detail fee not found")
			return errmsg.NewCustomErrors(404).SetMessage("Laporan tidak ditemukan")
		}
		log.Error().Err(err).Any("req", req).Msg("repo::UseHRfeeForLecturer - failed to get mentor detail fee")
		return err
	}

	if req.UsedAmount == nil && req.Notes == nil {
		log.Warn().Any("req", req).Msg("repo::UseHRfeeForLecturer - used amount and notes are nil")
		return errmsg.NewCustomErrors(400).SetMessage("Jumlah yang digunakan atau catatan harus diisi")
	}

	if req.UsedAmount != nil && req.Notes != nil {
		log.Warn().Any("req", req).Msg("repo::UseHRfeeForLecturer - used amount and notes are not nil")
		return errmsg.NewCustomErrors(400).SetMessage("Jumlah yang digunakan dan catatan tidak boleh diisi bersamaan")
	}

	if req.UsedAmount != nil && req.UsedAmount.GreaterThan(mentorDetailFee) {
		log.Warn().Any("req", req).Msg("repo::UseHRfeeForLecturer - used amount greater than mentor detail fee")
		return errmsg.NewCustomErrors(400).SetMessage("Jumlah yang digunakan melebihi jumlah yang tersedia")
	}

	if req.UsedAmount != nil && req.UsedAmount.LessThanOrEqual(decimal.Zero) {
		log.Warn().Any("req", req).Msg("repo::UseHRfeeForLecturer - used amount less than or equal to 0")
		return errmsg.NewCustomErrors(400).SetMessage("Jumlah yang digunakan harus lebih dari 0")
	}

	query := `
		UPDATE
			program_registrations
		SET
			mentor_detail_fee_used = ?,
			notes_for_fund_distributions = ?
		WHERE
			id = ?
			AND deleted_at IS NULL
	`

	_, err = Tx.ExecContext(ctx, Tx.Rebind(query), req.UsedAmount, req.Notes, req.RegistrationId)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UseHRfeeForLecturer - failed to update mentor detail fee used")
		return err
	}

	if err = Tx.Commit(); err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::UseHRfeeForLecturer - failed to commit transaction")
		return err
	}

	return nil
}
