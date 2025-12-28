package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (r *reportRepo) CreateAdditionalRegistration(ctx context.Context, req *entity.CreateAdditionalRegistrationReq) (*entity.CreateAdditionalRegistrationResp, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.CreateAdditionalRegistration")
	defer span.End()

	fnName := "repo::CreateAdditionalRegistration"

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("%s - failed to begin transaction", fnName)
		return nil, err
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Ctx(ctx).Error().Err(rbErr).Any("req", req).Msgf("%s - failed to rollback transaction", fnName)
			}
			return
		}
		if cmErr := tx.Commit(); cmErr != nil {
			log.Ctx(ctx).Error().Err(cmErr).Any("req", req).Msgf("%s - failed to commit transaction", fnName)
		}
	}()

	newID := ulid.Make().String()

	query := `
        WITH parent AS (
            SELECT
                template_id,
                program_id,
                marketer_id,
                student_id,
                lecturer_id,
                program_name,
                program_fee_per_meeting,
                program_meetings,
                program_meetings_completed,
                full_fee,
                marketer_gifts_fee,
                allocated_at,
                days
            FROM
                program_registrations pr
            WHERE
                pr.id = ?
        )
        INSERT INTO program_registrations (
            category,
            notes_for_category,
            parent_id,
            program_fee,

            id,
            template_id,
            user_id,
            program_id,
            marketer_id,
            student_id,
            lecturer_id,
            program_name,
            program_fee_per_meeting,
            program_meetings,
            program_meetings_completed,
            program_acquisition_rights,
            full_fee,
            marketer_commission_fee,
            hr_fee,
            hr_detail_fee,
            mentor_detail_fee,
            marketer_gifts_fee,
            closing_fee_for_office,
            closing_fee_for_reward,
            days,
            is_paid,
            paid_at,
            allocated_at,
            created_at,
            updated_at
        )
        SELECT
            ?,
            ?,
            ?,
            ?,

            ?,
            (SELECT template_id FROM parent),
            ?,
            (SELECT program_id FROM parent),
            (SELECT marketer_id FROM parent),
            (SELECT student_id FROM parent),
            (SELECT lecturer_id FROM parent),
            (SELECT program_name FROM parent),
            (SELECT program_fee_per_meeting FROM parent),
            (SELECT program_meetings FROM parent),
            (SELECT program_meetings_completed FROM parent),
            0,
            (SELECT full_fee FROM parent),
            ?,
            ?,
            0,
            ?,
            ?,

            ?,
            ?,

            (SELECT days FROM parent),
            TRUE,
            (? || ' ' || ?)::timestamp AT TIME ZONE 'Asia/Makassar',
            (SELECT allocated_at FROM parent),
            NOW(),
            NOW()
    `

	_, err = tx.ExecContext(ctx, tx.Rebind(query),
		req.ParentID,

		req.Category,
		req.NotesForCategory,
		req.ParentID,
		req.ProgramFee,

		newID,
		req.UserID,
		req.MarketerCommissionFee,
		req.HrFee,
		req.HrFee,
		req.MarketerGiftsFee,

		req.ClosingFeeForOffice,
		req.ClosingFeeForReward,

		req.PaidAt, req.PaidAtTime,
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to insert additional registration", fnName)
		return nil, err
	}

	return &entity.CreateAdditionalRegistrationResp{ID: newID}, nil
}
