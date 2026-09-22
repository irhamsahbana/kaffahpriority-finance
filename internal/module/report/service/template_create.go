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

func (s *reportService) CreateTemplate(ctx context.Context, req *entity.CreateTemplateReq) (*entity.CreateTemplateResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.CreateTemplate")
	defer span.End()

	resp, err := s.repo.CreateTemplate(ctx, req)
	if err != nil {
		return nil, err
	}

	childCtx := pkg.GenerateChildContext(ctx, time.Minute/2)
	go func() {
		user, err := s.userRepo.GetMe(childCtx, &entity.GetMeReq{UserID: req.UserID})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).Msgf("reportService.CreateTemplate - failed to get me")
			return
		}

		template, err := s.repo.GetTemplate(childCtx, &entity.GetTemplateReq{
			ID: resp.ID,
		})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).Msgf("reportService.CreateTemplate - failed to get template")
			return
		}

		err = s.activityLogRepo.CreateActivityLog(childCtx, &entity.ActivityLog{
			ID:         ulid.Make().String(),
			EntityName: entity.ActivityLogEntityProgramRegistrationTemplates,
			EntityID:   resp.ID,
			Type:       entity.ActivityLogTypeCreate,
			AuthorName: user.Name,
			AuthorRole: user.Role,
			CreatedAt:  time.Now(),
			CreatedBy:  req.UserID,
			Message:    "berhasil membuat template",
			After:      template,
		})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).Msgf("reportService.CreateTemplate - failed to create activity log")
		}
	}()

	return resp, nil
}
