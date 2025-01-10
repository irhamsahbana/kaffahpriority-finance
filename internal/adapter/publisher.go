package adapter

import (
	"codebase-app/internal/infrastructure/config"
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog/log"
)

func WithEmailNatsPublisher() Option {
	return func(a *Adapter) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		nc, err := nats.Connect(config.Envs.EmailVerificationQueueNats.NatsURL)
		if err != nil {
			log.Fatal().Err(err).Msg("Error while connecting to nats server")
		}

		js, err := jetstream.New(nc)
		if err != nil {
			log.Fatal().Err(err).Msg("Error while connecting to nats jetstream")
		}

		// create a stream
		_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
			Name:        "crowners-email-service",
			Description: "Email service stream",
			Subjects:    []string{"crowners.email.>"},
			MaxBytes:    1024 * 1024 * 1024,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("Error while creating nats jetstream stream")
		}

		a.EmailPublisher = js
		log.Info().Msg("Email NATS publisher connected")
	}
}

func WithExcelProductNatsPublisher() Option {
	return func(a *Adapter) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		nc, err := nats.Connect(config.Envs.EmailVerificationQueueNats.NatsURL)
		if err != nil {
			log.Fatal().Err(err).Msg("Error while connecting to nats server")
		}

		js, err := jetstream.New(nc)
		if err != nil {
			log.Fatal().Err(err).Msg("Error while connecting to nats jetstream")
		}

		// create a stream
		_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
			Name:        "excel-product-service",
			Description: "Excel product service stream",
			Subjects:    []string{"excel.>"},
			MaxBytes:    1024 * 1024 * 1024,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("Error while creating nats jetstream stream")
		}

		a.ExcelProductPublisher = js
		log.Info().Msg("Excel Product NATS publisher connected")
	}
}
