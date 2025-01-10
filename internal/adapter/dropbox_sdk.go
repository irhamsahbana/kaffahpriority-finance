package adapter

import (
	"codebase-app/internal/infrastructure/config"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/rs/zerolog/log"
)

func WithDropboxSDK() Option {
	return func(a *Adapter) {
		cfg := dropbox.Config{
			Token: config.Envs.Dropbox.AccessToken,
		}

		a.DropboxFiles = files.New(cfg)

		log.Info().Msg("Dropbox SDK initialized")
	}
}
