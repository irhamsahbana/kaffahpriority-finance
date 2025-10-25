package repository

import (
	"codebase-app/internal/module/report/entity"
	"codebase-app/pkg/errmsg"
	"context"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

func (r *reportRepo) UseHRfeeForLecturer(ctx context.Context, req *entity.UseHRfeeForLecturerReq) error {
	var (
		fnName = "repo::UseHRfeeForLecturer"
	)

	// Ambil semua registrasi terkait (termasuk dirinya sendiri)
	relatedResp, err := r.GetRelatedRegistrations(ctx, &entity.GetRelatedRegistrationsReq{
		RegistrationID: req.RegistrationID,
	})
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to get related registrations", fnName)
		return err
	}

	// Hitung ulang total fee & sisa fee tapi skip dirinya sendiri
	var (
		totalFee          decimal.Decimal
		totalFeeUsed      decimal.Decimal
		totalFeeRemaining decimal.Decimal
	)

	for _, item := range relatedResp.Items {
		totalFee = totalFee.Add(item.MentorDetailFee)
		totalFeeRemaining = totalFeeRemaining.Add(item.MentorDetailFee)

		// Skip diri sendiri agar validasi tidak mengacu ke dirinya
		if item.ID == req.RegistrationID {
			continue
		}

		var used decimal.Decimal
		if item.MentorDetailFeeUsed != nil {
			used = used.Add(*item.MentorDetailFeeUsed)
		}

		totalFeeUsed = totalFeeUsed.Add(used)
		totalFeeRemaining = totalFeeRemaining.Sub(used)
	}

	// Validasi: jumlah yang digunakan tidak boleh melebihi total fee tersisa dari related items
	if req.UsedAmount != nil && req.UsedAmount.GreaterThan(totalFeeRemaining) {
		errMsg := "jumlah yang diminta melebihi total sisa dana yang tersedia."
		log.Error().Any("req", req).Msgf("%s - %s", fnName, errMsg)
		return errmsg.
			NewCustomErrors(http.StatusUnprocessableEntity).
			Add("used_amount", errMsg).
			SetMessage(errMsg)
	}

	Tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to begin transaction", fnName)
		return err
	}
	defer Tx.Rollback()

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

	_, err = Tx.ExecContext(ctx, Tx.Rebind(query), req.UsedAmount, req.Notes, req.RegistrationID)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to update mentor detail fee used", fnName)
		return err
	}

	if err = Tx.Commit(); err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to commit transaction", fnName)
		return err
	}

	return nil
}

func (r *reportRepo) GetRelatedRegistrations(ctx context.Context, req *entity.GetRelatedRegistrationsReq) (*entity.GetRelatedRegistrationsResp, error) {
	var (
		fnName = "repo::GetRelatedRegistrations"
		resp   = new(entity.GetRelatedRegistrationsResp)
	)
	resp.Items = make([]entity.RelatedRegistration, 0)

	query := `
		WITH regis AS (
			SELECT
				pr.lecturer_id,
				pr.student_id,
				pr.program_id
			FROM
				program_registrations pr
			WHERE
				pr.deleted_at IS NULL
				AND
				pr.id = ?
		)
		SELECT
			pr.id,
			pr.lecturer_id,
			pr.program_id,
			pr.student_id,
			pr.mentor_detail_fee,
			pr.mentor_detail_fee_used,
			pr.paid_at,
			pr.allocated_at
		FROM
			program_registrations pr
		WHERE
			pr.deleted_at IS NULL
			AND
			pr.is_paid = TRUE
			AND
			pr.student_id = (SELECT student_id FROM regis)
			AND
			pr.program_id = (SELECT program_id FROM regis)
		ORDER BY
			pr.allocated_at ASC
	`

	err := r.db.SelectContext(ctx, &resp.Items, r.db.Rebind(query), req.RegistrationID)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to query related registrations", fnName)
		return nil, err
	}

	for _, item := range resp.Items {
		resp.TotalFee = resp.TotalFee.Add(item.MentorDetailFee)
		var feeUsed decimal.Decimal
		if item.MentorDetailFeeUsed != nil {
			feeUsed = feeUsed.Add(*item.MentorDetailFeeUsed)
		}
		resp.TotalFeeUsed = resp.TotalFeeUsed.Add(feeUsed)

		resp.TotalFeeRemaining = resp.TotalFeeRemaining.Add(item.MentorDetailFee).Sub(feeUsed)
	}

	return resp, nil
}
