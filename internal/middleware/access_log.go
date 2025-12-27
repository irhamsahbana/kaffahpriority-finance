package middleware

import (
	"codebase-app/internal/infrastructure"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func WithAccessLog(logger zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		if err != nil {
			return err
		}

		// get request id from context
		requestId, _ := c.Context().UserValue("request_id").(string)

		event := infrastructure.AccessLogger.Info().Ctx(c.UserContext()).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Any("query", c.Queries()).
			Str("ip", c.IP()).
			Str("user_agent", c.Get("User-Agent")).
			Dur("duration", time.Since(start)). // duration in ms
			Str("request_id", requestId)

		span := trace.SpanFromContext(c.UserContext())
		if span.SpanContext().IsValid() {
			event.Str("trace_id", span.SpanContext().TraceID().String())
			event.Str("span_id", span.SpanContext().SpanID().String())
			
			// Add access log event to span
			span.AddEvent("access_log", trace.WithAttributes(
				attribute.String("method", c.Method()),
				attribute.String("path", c.Path()),
				attribute.String("ip", c.IP()),
				attribute.Float64("duration_ms", float64(time.Since(start).Milliseconds())),
				attribute.String("request_id", requestId),
			))
		}

		event.Msg("access log")
		return nil
	}
}
