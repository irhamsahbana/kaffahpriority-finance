package tracing

import (
	"codebase-app/internal/infrastructure/config"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"
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
		resource.WithHost(),
		resource.WithOS(),
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
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	}

	if config.Envs.Instrumentation.Debug {
		exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout trace exporter: %w", err)
		}
		tracerProviderOptions = append(tracerProviderOptions, sdktrace.WithBatcher(exporter))
		log.Info().Msg("OpenTelemetry stdout exporter initialized (debug mode)")
	}

	otlpEndpoint := config.Envs.Instrumentation.OtlpEndpoint
	if otlpEndpoint != "" {
		var exporter sdktrace.SpanExporter
		var err error

		headers := parseHeaders(config.Envs.Instrumentation.OtlpHeaders)

		if strings.HasPrefix(otlpEndpoint, "http://") || strings.HasPrefix(otlpEndpoint, "https://") {
			// Use OTLP HTTP exporter
			// Strip scheme and trailing slash
			endpoint := strings.TrimPrefix(otlpEndpoint, "http://")
			endpoint = strings.TrimPrefix(endpoint, "https://")
			endpoint = strings.TrimRight(endpoint, "/")

			opts := []otlptracehttp.Option{
				otlptracehttp.WithEndpoint(endpoint),
			}

			if strings.HasPrefix(otlpEndpoint, "http://") {
				opts = append(opts, otlptracehttp.WithInsecure())
			}

			if len(headers) > 0 {
				opts = append(opts, otlptracehttp.WithHeaders(headers))
			}

			exporter, err = otlptracehttp.New(context.Background(), opts...)
			if err == nil {
				log.Info().Str("endpoint", otlpEndpoint).Msg("OpenTelemetry tracer initialized (OTLP HTTP exporter)")
			}
		} else {
			// Use OTLP gRPC exporter
			opts := []otlptracegrpc.Option{
				otlptracegrpc.WithEndpoint(otlpEndpoint),
			}

			if config.Envs.Instrumentation.OtlpInsecure {
				log.Info().Msg("Using Insecure connection for OTLP gRPC")
				opts = append(opts, otlptracegrpc.WithInsecure())
			} else {
				log.Info().Msg("Using Secure (TLS) connection for OTLP gRPC")
				// When connecting to public endpoints like New Relic (e.g. otlp.nr-data.net:4317),
				// we must use system certs but NOT WithInsecure().
				// WithTLSCredentials() enables TLS.
				// IMPORTANT: Do not mix WithInsecure() with WithTLSCredentials()

				// New Relic specifically requires TLS for gRPC on port 4317.
				// The error "frame too large, note that the frame header looked like an HTTP/1.1 header"
				// usually means we are sending non-TLS (HTTP/2 Cleartext) to a TLS-expecting server,
				// OR we are talking to an HTTP/1.1 server (like a proxy) instead of gRPC.

				// Ensure we are using system certs.
				// NOTE: We use &tls.Config{} to be explicit.
				// We also set InsecureSkipVerify to true ONLY for debugging if system roots are missing.
				// In production, this should be false.
				// Given the 'frame too large' error and openssl 'verify error: 20', it suggests a certificate issue.
				tlsConfig := &tls.Config{
					MinVersion: tls.VersionTLS12,
					// InsecureSkipVerify: true, // Uncomment if you have certificate issues (e.g. missing roots)
				}
				opts = append(opts, otlptracegrpc.WithTLSCredentials(credentials.NewTLS(tlsConfig)))
			}

			if len(headers) > 0 {
				log.Info().Int("header_count", len(headers)).Msg("Attaching OTLP headers")
				opts = append(opts, otlptracegrpc.WithHeaders(headers))
			}

			exporter, err = otlptracegrpc.New(context.Background(), opts...)
			if err == nil {
				log.Info().Str("endpoint", otlpEndpoint).Msg("OpenTelemetry tracer initialized (OTLP gRPC exporter)")
			}
		}

		if err != nil {
			return nil, fmt.Errorf("failed to create otlp exporter: %w", err)
		}
		tracerProviderOptions = append(tracerProviderOptions, sdktrace.WithBatcher(exporter))
	} else if !config.Envs.Instrumentation.Debug {
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

// StartSpan starts a new span using the global tracer
func StartSpan(ctx context.Context, name string, opts ...oteltrace.SpanStartOption) (context.Context, oteltrace.Span) {
	ctx, span := otel.Tracer(globalServiceName).Start(ctx, name, opts...)

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
			if k == "time" {
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
