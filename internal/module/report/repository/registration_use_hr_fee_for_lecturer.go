package repository

import (
	"codebase-app/internal/module/report/entity"
	"codebase-app/pkg/errmsg"
	"context"
	"fmt"
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
		// get lecturer, program and student
		registrationResp, err := r.GetRegistration(ctx, &entity.GetRegistrationReq{
			UserID: req.UserID,
			ID:     req.RegistrationID,
		})
		if err != nil {
			log.Error().Err(err).Any("req", req).Msgf("%s - failed to get registration", fnName)
			return err
		}

		var lecturerName string
		if registrationResp.LecturerName != nil {
			lecturerName = *registrationResp.LecturerName
		} else {
			lecturerName = "Mentor belum dipilih"
		}

		errMsg := "jumlah yang diminta melebihi total sisa dana yang tersedia."
		errMsg = fmt.Sprintf(
			"%s (mentor: %s, santri: %s, program: %s)",
			errMsg,
			lecturerName,
			registrationResp.StudentName,
			registrationResp.ProgramName,
		)

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

func (r *reportRepo) BulkUseHRfeeForLecturer(ctx context.Context, req *entity.BulkUseHRfeeForLecturerReq) error {
	var (
		fnName  = "repo::BulkUseHRfeeForLecturer"
		errBulk = []error{}
	)

	for _, item := range req.Data {
		err := r.UseHRfeeForLecturer(ctx, &item)
		if err != nil {
			log.Error().Err(err).Any("req", req).Any("item", item).Msgf("%s - failed to use hr fee for lecturer", fnName)
			errBulk = append(errBulk, err)
		}

	}

	if len(errBulk) > 0 {
		var errorMessages string
		for _, err := range errBulk {
			errorMessages += err.Error() + "\n"
		}

		return errmsg.NewCustomErrors(http.StatusUnprocessableEntity).
			SetMessage(errorMessages)
	}

	return nil
}
