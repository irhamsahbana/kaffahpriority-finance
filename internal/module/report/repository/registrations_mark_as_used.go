package repository

import (
	"codebase-app/internal/entity"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *reportRepo) RegistrationsMarkAsUsed(ctx context.Context, req *entity.RegistrationsMarkAsUsedReq) error {
	fnName := "repo::RegistrationsMarkAsUsed"

	query := `
        UPDATE program_registrations
            SET mentor_detail_fee_used = mentor_detail_fee,
			updated_at = now()
        WHERE
            deleted_at IS NULL
			AND mentor_detail_fee_used IS NULL
			AND category = 'general'
			AND is_paid = true
			AND program_meetings > 0
            AND allocated_at AT TIME ZONE 'Asia/Makassar' >=
            (TO_TIMESTAMP(?, 'YYYY-MM') AT TIME ZONE 'UTC')
            AND allocated_at AT TIME ZONE 'Asia/Makassar' <
            (TO_TIMESTAMP(?, 'YYYY-MM') AT TIME ZONE 'UTC' + INTERVAL '1 month')
    `

	if _, err := r.db.ExecContext(ctx, r.db.Rebind(query), req.AllocatedMonth, req.AllocatedMonth); err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to update data", fnName)
		return err
	}

	return nil
}
