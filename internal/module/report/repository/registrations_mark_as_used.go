package repository

import (
	"codebase-app/internal/module/report/entity"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *reportRepo) RegistrationsMarkAsUsed(ctx context.Context, req *entity.RegistrationsMarkAsUsedReq) error {
	fnName := "repo::RegistrationsMarkAsUsed"

	query := `
        UPDATE program_registrations
            SET mentor_detail_fee_used = mentor_detail_fee
        WHERE
            deleted_at IS NULL
			AND
			category = 'general'
			AND
			program_meetings > 0
            AND
            allocated_at >= (TO_TIMESTAMP(?, 'YYYY-MM') AT TIME ZONE 'Asia/Makassar')
            AND allocated_at < ((TO_TIMESTAMP(?, 'YYYY-MM') AT TIME ZONE 'Asia/Makassar') + INTERVAL '1 month')
    `

	if _, err := r.db.ExecContext(ctx, r.db.Rebind(query), req.AllocatedMonth, req.AllocatedMonth); err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to update data", fnName)
		return err
	}

	return nil
}
