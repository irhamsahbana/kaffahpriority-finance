package adapter

import (
	"codebase-app/internal/infrastructure/config"

	"github.com/rs/zerolog/log"
)

func WithDropboxSDK() Option {
	return func(a *Adapter) {
		// Initialize token manager
		tokenManager := NewDropboxTokenManager(
			config.Envs.Dropbox.AccessToken,
			config.Envs.Dropbox.RefreshToken,
			config.Envs.Dropbox.AppKey,
			config.Envs.Dropbox.AppSecret,
		)

		// Create dynamic client with automatic token refresh
		dynamicClient := NewDynamicDropboxClient(tokenManager)

		a.DropboxFiles = dynamicClient
		a.DropboxTokenManager = tokenManager

		log.Info().Msg("Dropbox SDK initialized with OAuth2 token manager and dynamic client")
	}
}
