package adapter

import (
	"context"

	firebase "firebase.google.com/go"
	"github.com/rs/zerolog/log"

	"google.golang.org/api/option"
)

func WithFirebaseSDK() Option {
	return func(a *Adapter) {
		opt := option.WithCredentialsFile("./serviceAccountKey.json")
		app, err := firebase.NewApp(context.Background(), nil, opt)
		if err != nil {
			log.Fatal().Err(err).Msg("Error initializing firebase app")
		}

		a.FirebaseSDK = app
	}
}
