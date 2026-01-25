package cmd

import (
	"codebase-app/internal/adapter"
	"codebase-app/internal/infrastructure/config"
	"codebase-app/internal/infrastructure/logging"
	"context"
	"flag"
	"os"
	"os/signal"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog/log"
)

func RunConsumer(_ *flag.FlagSet, _ []string) {
	envs := config.Envs

	var otlpEndpoint string
	if envs.Instrumentation.Enabled {
		otlpEndpoint = envs.Instrumentation.OtlpEndpoint
	}

	lp, _, err := logging.InitLogger(&logging.Config{
		Endpoint:   otlpEndpoint,
		AppName:    envs.App.Name,
		AppVersion: envs.App.Version,
		AppEnv:     envs.App.Environtment,
		LogFile:    "consumer.log",
		LogLevel:   envs.App.LogLevel,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize logger")
	}
	defer func() {
		_ = lp.Shutdown(context.Background())
	}()

	log.Info().Msg("Running consumer")
	var (
		emailConsumerCtx jetstream.ConsumeContext
	)

	adapter.Adapters.Sync(
		adapter.WithEmailConsumerNats(emailConsumerCtx),
	)

	EmailConsumer := adapter.Adapters.EmailConsumerNats

	// email consumer
	cctxEmail, err := EmailConsumer.Consume(func(msg jetstream.Msg) {
		switch msg.Subject() {
		case "email.verification":
			EmailVerificationHandler(msg)
		case "email.forgot-password":
			ForgotPasswordHandler(msg)
		default:
		}
	})
	if err != nil {
		log.Fatal().Err(err).Msg("consumer::RunConsumer::Error while consuming message")
	}

	adapter.Adapters.EmailConsumerCtxNats = cctxEmail

	defer func() {
		if err := adapter.Adapters.Unsync(); err != nil {
			log.Fatal().Err(err).Msg("Error while closing database connection")
		}
	}()

	// gracefully shutdown the consumer
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Info().Msg("Consumer gracefully stopped")
}
