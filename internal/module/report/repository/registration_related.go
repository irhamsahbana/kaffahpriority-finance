package repository

import (
	"codebase-app/internal/entity"
	"context"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

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
			pr.category,
			pr.lecturer_id,
			pr.program_id,
			pr.student_id,
			pr.mentor_detail_fee,
			pr.mentor_detail_fee_used,
			pr.paid_at,
			COALESCE(pr.allocated_at, p_p.allocated_at) as allocated_at
		FROM
			program_registrations pr
		LEFT JOIN
			program_registrations p_p ON pr.parent_id = p_p.id
		WHERE
			pr.deleted_at IS NULL
			AND
			pr.is_paid = TRUE
			AND
			pr.student_id = (SELECT student_id FROM regis)
			AND
			pr.program_id = (SELECT program_id FROM regis)
			AND
			pr.mentor_detail_fee > 0
		ORDER BY
			allocated_at ASC
	`

	err := r.db.SelectContext(ctx, &resp.Items, r.db.Rebind(query), req.RegistrationID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to query related registrations", fnName)
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
