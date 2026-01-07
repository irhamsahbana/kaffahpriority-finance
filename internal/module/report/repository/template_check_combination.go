package repository

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"context"

	"github.com/rs/zerolog/log"
)

func (r *reportRepo) CheckTemplateCombinationForUpdate(ctx context.Context, req *entity.UpdateTemplateGeneralReq) (bool, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.CheckTemplateCombinationForUpdate")
	defer span.End()

	fnName := "repo::CheckTemplateCombinationForUpdate"

	var isCombinationExist bool
	queryCombination := `
		SELECT EXISTS (
			SELECT 1
			FROM program_registration_templates prt
			WHERE
				prt.program_id = ?
				AND prt.marketer_id = ?
				AND prt.student_id = ?
				AND (
					(prt.lecturer_id IS NULL AND ?::TEXT IS NULL)
					OR prt.lecturer_id = ?
				)
				AND prt.id != ?
				AND prt.deleted_at IS NULL
		)
	`

	err := r.db.GetContext(ctx, &isCombinationExist, r.db.Rebind(queryCombination),
		req.ProgramId, req.MarketerId, req.StudentId, req.LecturerId, req.LecturerId, req.ID,
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("%s - failed to check combination", fnName)
		return false, err
	}

	return isCombinationExist, nil
}

func (r *reportRepo) CheckTemplateCombinationForCreate(ctx context.Context, req *entity.CreateTemplateReq) (bool, error) {
	if req.LecturerID == nil {
		return r.CheckTemplateCombinationForCreateIfLecturerNotExist(ctx, req)
	} else {
		return r.CheckTemplateCombinationForCreateIfLecturerExist(ctx, req)
	}
}

func (r *reportRepo) CheckTemplateCombinationForCreateIfLecturerNotExist(ctx context.Context, req *entity.CreateTemplateReq) (bool, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.CheckTemplateCombinationForCreateIfLecturerNotExist")
	defer span.End()

	var isCombinationExist bool
	queryCombination := `
		SELECT EXISTS (
			SELECT 1
			FROM program_registration_templates prt
			WHERE
				prt.program_id = ?
				AND prt.marketer_id = ?
				AND prt.student_id = ?
				AND prt.lecturer_id IS NULL
				AND prt.deleted_at IS NULL
		)
	`

	err := r.db.GetContext(ctx, &isCombinationExist, r.db.Rebind(queryCombination),
		req.ProgramID, req.MarketerID, req.StudentID,
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to check combination")
		return false, err
	}

	return isCombinationExist, nil
}

func (r *reportRepo) CheckTemplateCombinationForCreateIfLecturerExist(ctx context.Context, req *entity.CreateTemplateReq) (bool, error) {
	ctx, span := tracing.StartSpan(ctx, "repo.CheckTemplateCombinationForCreateIfLecturerExist")
	defer span.End()

	var isCombinationExist bool
	queryCombination := `
		SELECT EXISTS (
			SELECT 1
			FROM program_registration_templates prt
			WHERE
				prt.program_id = ?
				AND prt.student_id = ?
				AND prt.marketer_id = ?
				AND prt.lecturer_id = ?
				AND prt.deleted_at IS NULL
		)
	`

	err := r.db.GetContext(ctx, &isCombinationExist, r.db.Rebind(queryCombination),
		req.ProgramID, req.StudentID, req.MarketerID, *req.LecturerID,
	)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Any("req", req).Msgf("failed to check combination")
		return false, err
	}

	return isCombinationExist, nil
}
