package tracing

import (
	"codebase-app/internal/infrastructure/config"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var globalServiceName string
var globalLogWriter io.Writer

type Config struct {
	Endpoint   string
	AppName    string
	AppVersion string
	AppEnv     string
	LogWriter  io.Writer
}

// InitTracer initializes the OpenTelemetry tracer provider
func InitTracer(cfg *Config) (*sdktrace.TracerProvider, error) {
	globalServiceName = cfg.AppName
	globalLogWriter = cfg.LogWriter

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.AppName),
			semconv.DeploymentEnvironment(cfg.AppEnv),
			semconv.ServiceVersion(cfg.AppVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tracerProviderOptions := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
	}

	switch cfg.AppEnv {
	case "development":
		tracerProviderOptions = append(tracerProviderOptions, sdktrace.WithSampler(sdktrace.AlwaysSample())) // Sample all traces for dev/demo
	case "production":
		tracerProviderOptions = append(tracerProviderOptions, sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.1)))) // Sample 10% of traces for production
	}

	otlpEndpoint := config.Envs.Instrumentation.OtlpEndpoint
	if otlpEndpoint != "" {
		// Use OTLP gRPC exporter
		exporter, err := otlptracegrpc.New(context.Background(),
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(otlpEndpoint),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create otlp exporter: %w", err)
		}
		log.Info().Str("endpoint", otlpEndpoint).Msg("OpenTelemetry tracer initialized (OTLP gRPC exporter)")
		tracerProviderOptions = append(tracerProviderOptions, sdktrace.WithBatcher(exporter))
	} else {
		log.Info().Msg("OpenTelemetry tracer initialized (no exporter, tracing disabled)")
	}

	tp := sdktrace.NewTracerProvider(tracerProviderOptions...)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

// StartSpan starts a new span using the global tracer
func StartSpan(ctx context.Context, name string) (context.Context, oteltrace.Span) {
	ctx, span := otel.Tracer(globalServiceName).Start(ctx, name)

	// Inject logger with trace_id and span_id into context
	if sc := span.SpanContext(); sc.IsValid() {
		var output io.Writer = os.Stderr
		if globalLogWriter != nil {
			output = globalLogWriter
		}

		spanWriter := &SpanLogWriter{Next: output, Span: span}

		// Start with a fresh logger to avoid inheriting hooks from parent spans
		// Use Output() to replace the writer with our interceptor
		logger := log.Logger.Output(spanWriter).With().
			Str("trace_id", sc.TraceID().String()).
			Str("span_id", sc.SpanID().String()).
			Logger()

		// Re-inject request_id if present in context
		if reqID, ok := ctx.Value("request_id").(string); ok {
			logger = logger.With().Str("request_id", reqID).Logger()
		}

		ctx = logger.WithContext(ctx)
	}

	return ctx, span
}

type SpanLogWriter struct {
	Next io.Writer
	Span oteltrace.Span
}

func (w *SpanLogWriter) Write(p []byte) (n int, err error) {
	n, err = w.Next.Write(p)
	if err != nil {
		return n, err
	}

	var data map[string]interface{}
	if json.Unmarshal(p, &data) == nil {
		attrs := []attribute.KeyValue{}
		for k, v := range data {
			if k == "time" || k == "span_id" || k == "trace_id" {
				continue
			}
			if strVal, ok := v.(string); ok {
				attrs = append(attrs, attribute.String("log."+k, strVal))
			} else {
				// Try to marshal complex types (maps, slices, structs) to JSON string
				// This avoids "map[key:value]" formatting in traces
				if jsonBytes, err := json.Marshal(v); err == nil {
					attrs = append(attrs, attribute.String("log."+k, string(jsonBytes)))
				} else {
					attrs = append(attrs, attribute.String("log."+k, fmt.Sprintf("%v", v)))
				}
			}
		}
		w.Span.AddEvent("log", oteltrace.WithAttributes(attrs...))

		if lvl, ok := data["level"].(string); ok {
			if lvl == "error" || lvl == "fatal" || lvl == "panic" {
				if _, hasErr := data["error"]; !hasErr {
					if msg, ok := data["message"].(string); ok {
						w.Span.RecordError(errors.New(msg))
						w.Span.SetStatus(codes.Error, msg)
					}
				}
			}
			if lvl == "warn" {
				w.Span.SetAttributes(attribute.Bool("has_warning", true))
			}
		}
	}
	return n, nil
}
