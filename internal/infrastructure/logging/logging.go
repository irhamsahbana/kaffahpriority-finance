package logging

import (
	"codebase-app/internal/infrastructure/config"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/log/global"
	logsdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"google.golang.org/grpc/credentials"
	"gopkg.in/natefinch/lumberjack.v2"
)

// AccessLogger is a dedicated logger for access logs.
var AccessLogger zerolog.Logger

type Config struct {
	Endpoint      string
	AppName       string
	AppVersion    string
	AppEnv        string
	LogFile       string
	AccessLogFile string
	LogLevel      string
}

func InitLogger(cfg *Config) (*logsdk.LoggerProvider, io.Writer, error) {
	// 1. Initialize OTEL Provider
	provider, err := initOtelProvider(cfg)
	if err != nil {
		return nil, nil, err
	}

	// 2. Initialize Zerolog (App Logger)
	writer := initZerolog(cfg)

	// 3. Initialize Access Logger
	if cfg.AccessLogFile != "" {
		initAccessLogger(cfg)
	}

	return provider, writer, nil
}

func initOtelProvider(cfg *Config) (*logsdk.LoggerProvider, error) {
	res, err := resource.New(context.Background(),
		resource.WithHost(),
		resource.WithOS(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.AppName),
			attribute.String("deployment.environment", cfg.AppEnv),
			semconv.ServiceVersionKey.String(cfg.AppVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	var loggerProviderOptions []logsdk.LoggerProviderOption
	loggerProviderOptions = append(loggerProviderOptions, logsdk.WithResource(res))

	if config.Envs.Instrumentation.Debug {
		exporter, err := stdoutlog.New(stdoutlog.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout log exporter: %w", err)
		}
		loggerProviderOptions = append(loggerProviderOptions, logsdk.WithProcessor(logsdk.NewBatchProcessor(exporter)))
		log.Info().Msg("OpenTelemetry stdout log exporter initialized (debug mode)")
	}

	otlpEndpoint := cfg.Endpoint
	if otlpEndpoint != "" {
		var exporter logsdk.Exporter
		var err error

		if strings.HasPrefix(otlpEndpoint, "http://") || strings.HasPrefix(otlpEndpoint, "https://") {
			// Use OTLP HTTP exporter
			endpoint := strings.TrimPrefix(otlpEndpoint, "http://")
			endpoint = strings.TrimPrefix(endpoint, "https://")
			endpoint = strings.TrimRight(endpoint, "/")

			opts := []otlploghttp.Option{
				otlploghttp.WithEndpoint(endpoint),
			}
			if strings.HasPrefix(otlpEndpoint, "http://") {
				opts = append(opts, otlploghttp.WithInsecure())
			}

			exporter, err = otlploghttp.New(context.Background(), opts...)
			if err == nil {
				log.Info().Str("endpoint", otlpEndpoint).Msg("OpenTelemetry logger initialized (OTLP HTTP exporter)")
			}
		} else {
			// Use OTLP gRPC exporter
			opts := []otlploggrpc.Option{
				otlploggrpc.WithEndpoint(otlpEndpoint),
			}

			if config.Envs.Instrumentation.OtlpInsecure {
				opts = append(opts, otlploggrpc.WithInsecure())
			} else {
				tlsConfig := &tls.Config{
					MinVersion: tls.VersionTLS12,
				}
				opts = append(opts, otlploggrpc.WithTLSCredentials(credentials.NewTLS(tlsConfig)))
			}

			headers := parseHeaders(config.Envs.Instrumentation.OtlpHeaders)
			if len(headers) > 0 {
				opts = append(opts, otlploggrpc.WithHeaders(headers))
			}

			exporter, err = otlploggrpc.New(context.Background(), opts...)
			if err == nil {
				log.Info().Str("endpoint", otlpEndpoint).Msg("OpenTelemetry logger initialized (OTLP gRPC exporter)")
			}
		}

		if err != nil {
			return nil, fmt.Errorf("failed to create otlp log exporter: %w", err)
		}
		loggerProviderOptions = append(loggerProviderOptions, logsdk.WithProcessor(logsdk.NewBatchProcessor(exporter)))
	} else if !config.Envs.Instrumentation.Debug {
		log.Info().Msg("OpenTelemetry logger initialized (no exporter, logging disabled)")
	}

	// LoggerProvider with batcher
	provider := logsdk.NewLoggerProvider(loggerProviderOptions...)

	// Set global LoggerProvider
	global.SetLoggerProvider(provider)
	return provider, nil
}

func initZerolog(cfg *Config) io.Writer {
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	var (
		lumberjackLogger = &lumberjack.Logger{
			MaxSize:  100, // megabytes
			MaxAge:   1,   // days
			Filename: cfg.LogFile,
		}
		writers = []io.Writer{zerolog.ConsoleWriter{Out: os.Stderr}, lumberjackLogger}
		mw      = io.MultiWriter(writers...)
	)

	var baseWriter io.Writer
	if cfg.AppEnv == "production" {
		baseWriter = lumberjackLogger
	} else {
		baseWriter = mw
	}

	// Wrap with OtelWriter
	provider := global.GetLoggerProvider()
	otelLogger := provider.Logger("codebase-app")
	otelWriter := &OtelWriter{
		Next:   baseWriter,
		Logger: otelLogger,
	}

	var logger zerolog.Logger
	if cfg.AppEnv == "production" {
		logger = zerolog.New(otelWriter).With().Timestamp().Caller().Logger().Level(zerolog.InfoLevel)
	} else {
		logger = zerolog.New(otelWriter).With().Timestamp().Caller().Logger().Level(level)
	}
	log.Logger = logger

	// Signal handling for rotation
	handleLogRotation(lumberjackLogger)

	return otelWriter
}

func initAccessLogger(cfg *Config) {
	// ensure directory exists
	if err := os.MkdirAll(filepath.Dir(cfg.AccessLogFile), os.ModePerm); err != nil {
		log.Fatal().Err(err).Msg("failed to create access log directory")
	}

	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	var (
		lumberjackLogger = &lumberjack.Logger{
			MaxSize:  100, // megabytes
			MaxAge:   1,   // days
			Filename: cfg.AccessLogFile,
		}
		writers = []io.Writer{zerolog.ConsoleWriter{Out: os.Stderr}, lumberjackLogger}
		mw      = io.MultiWriter(writers...)
	)

	var baseWriter io.Writer
	if cfg.AppEnv == "production" {
		baseWriter = lumberjackLogger
	} else {
		baseWriter = mw
	}

	// Wrap with OtelWriter
	provider := global.GetLoggerProvider()
	otelLogger := provider.Logger("codebase-app.access")
	otelWriter := &OtelWriter{
		Next:   baseWriter,
		Logger: otelLogger,
	}

	var logger zerolog.Logger
	if cfg.AppEnv == "production" {
		logger = zerolog.New(otelWriter).With().Timestamp().Caller().Logger().Level(zerolog.InfoLevel)
	} else {
		logger = zerolog.New(otelWriter).With().Timestamp().Caller().Logger().Level(level)
	}
	AccessLogger = logger

	// Signal handling for rotation
	handleLogRotation(lumberjackLogger)
}

func handleLogRotation(l *lumberjack.Logger) {
	q := make(chan os.Signal, 1)
	c := make(chan os.Signal, 1)
	signal.Notify(q, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)
	signal.Notify(c, syscall.SIGHUP)
	go func() {
		for {
			<-q
			l.Close()
			log.Info().Msg("Closing logs ...")
		}
	}()
	go func() {
		for {
			<-c
			if err := l.Rotate(); err != nil {
				log.Error().Err(err).Msg("Error while rotating logs")
			}
			log.Info().Msg("Rotating logs ...")
		}
	}()
}

func parseHeaders(headersStr string) map[string]string {
	headers := make(map[string]string)
	if headersStr == "" {
		return headers
	}
	pairs := strings.Split(headersStr, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return headers
}
