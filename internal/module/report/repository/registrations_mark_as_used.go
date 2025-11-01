package repository

import (
	"codebase-app/internal/module/report/entity"
	"context"

	"github.com/jmoiron/sqlx"
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
            mentor_detail_fee_used IS NOT NULL
            AND
            id IN (?)
    `

	q, args, err := sqlx.In(query, req.RegistrationIds)
	if err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to build query", fnName)
		return err
	}

	q = r.db.Rebind(q)
	if _, err = r.db.ExecContext(ctx, q, args...); err != nil {
		log.Error().Err(err).Any("req", req).Msgf("%s - failed to update data", fnName)
		return err
	}

	return nil
}
