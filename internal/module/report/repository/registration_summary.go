package repository

import (
	"codebase-app/internal/module/report/entity"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *reportRepo) GetSummaries(ctx context.Context, req *entity.GetSummariesReq) (*entity.GetSummariesResp, error) {
	var (
		resp = new(entity.GetSummariesResp)
		args = make([]any, 0, 3)
	)

	query := `
		SELECT
			COALESCE(SUM(pr.hr_fee), 0) AS total_hr_fee,
			COALESCE(SUM(pr.overpayment_fee), 0) AS total_overpayment_fee,
			COALESCE(SUM(pr.marketer_commission_fee), 0) AS total_marketer_commission_fee,
			COALESCE(SUM(pr.marketer_gifts_fee), 0) AS total_marketer_gifts_fee,
			COALESCE(SUM(pr.closing_fee_for_reward), 0) AS total_closing_fee_for_reward,
			COALESCE(
				SUM(
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
				)
			, 0) AS total_profit
		FROM
			program_registrations pr
		WHERE
			pr.deleted_at IS NULL
			AND pr.paid_at AT TIME ZONE ? BETWEEN
			(TO_TIMESTAMP(?, 'YYYY-MM-DD') AT TIME ZONE 'UTC') AND
			(TO_TIMESTAMP(?, 'YYYY-MM-DD') AT TIME ZONE 'UTC' + time '23:59:59.999999')
	`
	args = append(args, req.Timezone, req.PaidAtFrom, req.PaidAtTo)

	err := r.db.QueryRowContext(ctx, r.db.Rebind(query), args...).Scan(
		&resp.TotalHrFee,
		&resp.TotalOverpaymentFee,
		&resp.TotalMarketerCommission,
		&resp.TotalMarketerGifts,
		&resp.TotalClosingFeeForReward,
		&resp.TotalProfit,
	)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msg("repo::GetSummaries - failed to get summaries")
		return nil, err
	}

	resp.PaidAtFrom = req.PaidAtFrom
	resp.PaidAtTo = req.PaidAtTo

	return resp, nil
}
