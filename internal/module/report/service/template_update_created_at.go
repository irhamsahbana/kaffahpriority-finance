package service

import (
	"codebase-app/internal/entity"
	"codebase-app/internal/infrastructure/tracing"
	"codebase-app/pkg"
	"codebase-app/pkg/errmsg"
	"context"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
)

func (s *reportService) UpdateTemplateCreatedAtBetween(ctx context.Context, req *entity.UpdateTemplateCreatedAtBetweenReq) error {
	ctx, span := tracing.StartSpan(ctx, "service.UpdateTemplateCreatedAtBetween")
	defer span.End()

	var (
		prevTime, nextTime, newTime time.Time
		err                         error
	)

	if req.PrevID != "" {
		prevTime, err = s.repo.GetTemplateCreatedAt(ctx, req.PrevID)
		if err != nil {
			return err
		}
	}

	if req.NextID != "" {
		nextTime, err = s.repo.GetTemplateCreatedAt(ctx, req.NextID)
		if err != nil {
			return err
		}
	}

	if req.PrevID != "" && req.NextID != "" {
		// Case 1: Insert between two items
		duration := nextTime.Sub(prevTime)
		newTime = prevTime.Add(duration / 2)
	} else if req.PrevID == "" && req.NextID != "" {
		// Case 2: Insert at the beginning (before NextID)
		newTime = nextTime.Add(-30 * time.Minute)
	} else if req.PrevID != "" && req.NextID == "" {
		// Case 3: Insert at the end (after PrevID)
		newTime = prevTime.Add(30 * time.Minute)
	} else {
		// Case 4: Both missing (Invalid)
		return errmsg.NewCustomErrors(400).SetMessage("At least one of prev_id or next_id must be provided")
	}

	err = s.repo.UpdateTemplateCreatedAt(ctx, req.TargetID, newTime)
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
		Time("new_target_time", newTime).
		Msg("updated template created_at between")

	return nil
}
