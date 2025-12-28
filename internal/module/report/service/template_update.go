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

func (s *reportService) UpdateTemplate(ctx context.Context, req *entity.UpdateTemplateGeneralReq) (*entity.UpdateTemplateResp, error) {
	ctx, span := tracing.StartSpan(ctx, "service.UpdateTemplate")
	defer span.End()

	isCombinationExist, err := s.repo.CheckTemplateCombinationForUpdate(ctx, req)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).
			Any(entity.Payload, req).
			Msgf("failed to check combination")
		return nil, err
	}
	if isCombinationExist {
		log.Ctx(ctx).Warn().
			Any(entity.Payload, req).
			Msgf("combination already exist")
		return nil, errmsg.NewCustomErrors(409).SetMessage("Template dengan kombinasi program, marketer, pengajar, dan santri tersebut sudah ada. Silahkan cek kembali atau update data yang sudah ada")
	}

	oldTemplate, err := s.repo.GetTemplate(ctx, &entity.GetTemplateReq{
		ID: req.ID,
	})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).
			Any(entity.Payload, req).
			Msgf("failed to get template")
		return nil, err
	}

	resp, err := s.repo.UpdateTemplate(ctx, req)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).
			Any(entity.Payload, req).
			Msgf("failed to update template")
		return nil, err
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
			ID: resp.ID,
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
			EntityID:   resp.ID,
			Type:       entity.ActivityLogTypeUpdate,
			AuthorName: user.Name,
			AuthorRole: user.Role,
			UpdatedAt:  time.Now(),
			UpdatedBy:  req.UserID,
			Message:    "berhasil mengupdate template",
			Before:     oldTemplate,
			After:      template,
		})
		if err != nil {
			log.Ctx(childCtx).Error().Err(err).
				Any(entity.Payload, req).
				Msgf("failed to create activity log")
		}
	}()

	return resp, nil
}
