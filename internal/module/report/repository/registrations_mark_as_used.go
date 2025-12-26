package repository

import (
	"codebase-app/internal/entity"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *reportRepo) RegistrationsMarkAsUsed(ctx context.Context, req *entity.RegistrationsMarkAsUsedReq) ([]string, error) {
	fnName := "repo::RegistrationsMarkAsUsed"

	query := `
        UPDATE program_registrations pr
        SET 
            mentor_detail_fee_used = pr.mentor_detail_fee,
            updated_at = now()
        FROM (
            SELECT 
                child.id
            FROM 
                program_registrations child
            LEFT JOIN 
                program_registrations parent ON child.parent_id = parent.id
            WHERE 
                child.deleted_at IS NULL
                AND child.mentor_detail_fee_used IS NULL
                AND child.is_paid = true
                AND child.category IN ('general', 'additional')
                AND COALESCE(child.program_meetings, parent.program_meetings, 0) > 0
                AND TO_CHAR(COALESCE(child.allocated_at, parent.allocated_at) AT TIME ZONE 'Asia/Makassar', 'YYYY-MM') = ?
        ) as sub
        WHERE pr.id = sub.id
		RETURNING pr.id
    `

	var ids []string
	if err := r.db.SelectContext(ctx, &ids, r.db.Rebind(query), req.AllocatedMonth); err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to update data", fnName)
		return nil, err
	}

	return ids, nil
}
