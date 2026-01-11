package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/pkg"
	"context"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (s *reportService) UpdateTemplateCreatedAtBetween(ctx context.Context, req *entity.UpdateTemplateCreatedAtBetweenReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.UpdateTemplateCreatedAtBetween")
	defer span.End()

	prevTime, err := s.repo.GetTemplateCreatedAt(ctx, req.PrevID)
	if err != nil {
		return err
	}

	nextTime, err := s.repo.GetTemplateCreatedAt(ctx, req.NextID)
	if err != nil {
		return err
	}

	// Calculate midpoint
	duration := nextTime.Sub(prevTime)
	midPoint := prevTime.Add(duration / 2)

	err = s.repo.UpdateTemplateCreatedAt(ctx, req.TargetID, midPoint)
	if err != nil {
		return err
	}

	childCtx := pkg.GenerateChildContext(ctx, time.Minute/2)
	go func() {
		user, err := s.userRepo.GetMe(childCtx, &entity.GetMeReq{UserID: req.UserID})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get me")
			return
		}

		template, err := s.repo.GetTemplate(childCtx, &entity.GetTemplateReq{
			UserID: req.UserID,
			ID:     req.TargetID,
		})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to get template")
			return
		}

		err = s.activityLogRepo.CreateActivityLog(childCtx, &entity.ActivityLog{
			ID:         ulid.Make().String(),
			EntityName: entity.ActivityLogEntityProgramRegistrationTemplates,
			EntityID:   req.TargetID,
			Type:       entity.ActivityLogTypeUpdate,
			AuthorName: user.Name,
			AuthorRole: user.Role,
			UpdatedAt:  time.Now(),
			UpdatedBy:  req.UserID,
			Message:    "berhasil mengupdate created_at template (sisipan)",
			After:      template,
		})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to create activity log")
		}
	}()

	log.Ctx(ctx).Info().
		Str("prev_id", req.PrevID).
		Str("next_id", req.NextID).
		Str("target_id", req.TargetID).
		Time("prev_time", prevTime).
		Time("next_time", nextTime).
		Time("new_target_time", midPoint).
		Msg("updated template created_at between")

	return nil
}
