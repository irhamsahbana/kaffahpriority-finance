package adapter

import (
	"codebase-app/internal/infrastructure/config"
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog/log"
)

func WithEmailConsumerNats(cctx jetstream.ConsumeContext) Option {
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
			Description: "Email service stream for crowners app",
			Subjects:    []string{"crowners.email.>"},
			MaxBytes:    1024 * 1024 * 1024,  // 1GB
			MaxAge:      time.Hour * 24 * 14, // 14 days
		})
		if err != nil {
			log.Fatal().Err(err).Msg("Error while creating nats jetstream stream")
		}

		stream, err := js.Stream(ctx, "crowners-email-service")
		if err != nil {
			log.Fatal().Err(err).Msg("Error while getting nats jetstream stream")
		}

		consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
			Name:        "crowners-email-service-consumer",
			Durable:     "crowners-email-service-consumer",
			Description: "Email service consumer for crowners app",
		})
		if err != nil {
			log.Fatal().Err(err).Msg("Error while creating nats jetstream consumer")
		}

		a.EmailConsumerNats = consumer
		a.EmailConsumerCtxNats = cctx
	}
}
