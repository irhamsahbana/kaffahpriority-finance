package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// TracingMiddleware returns a Fiber handler that instruments requests with OpenTelemetry
func WithTracing(serviceName string) fiber.Handler {
	tracer := otel.Tracer(serviceName)

	return func(c *fiber.Ctx) error {
		// Manual extraction because HeaderCarrier expects http.Header
		header := make(propagation.HeaderCarrier)
		for k, v := range c.GetReqHeaders() {
			header[k] = v
		}

		ctx := otel.GetTextMapPropagator().Extract(c.UserContext(), header)

		// Use c.Path() initially as it contains the actual request path
		spanName := c.Method() + " " + c.Path()
		ctx, span := tracer.Start(ctx, spanName,
			oteltrace.WithAttributes(
				semconv.HTTPMethod(c.Method()),
				semconv.HTTPTarget(c.Path()),
				semconv.HTTPURL(c.OriginalURL()),
				semconv.HTTPClientIP(c.IP()),
				attribute.String("user_agent.original", c.Get("User-Agent")),
			),
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		)
		defer span.End()

		// Inject logger with trace_id and span_id into context
		if sc := span.SpanContext(); sc.IsValid() {
			logger := log.With().
				Str("trace_id", sc.TraceID().String()).
				Str("span_id", sc.SpanID().String()).
				Logger()
			ctx = logger.WithContext(ctx)
		}

		c.SetUserContext(ctx)

		if sc := span.SpanContext(); sc.HasTraceID() {
			c.Set("X-Trace-ID", sc.TraceID().String())
		}

		err := c.Next()

		// Update span name with the matched route path to ensure low cardinality
		// e.g. /users/123 -> /users/:id
		routePath := c.Route().Path
		if routePath != "" && routePath != "/" {
			span.SetName(c.Method() + " " + routePath)
			span.SetAttributes(semconv.HTTPRoute(routePath))
		}


		status := c.Response().StatusCode()
		span.SetAttributes(semconv.HTTPStatusCode(status))

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			if status >= 500 {
				span.SetStatus(codes.Error, "internal server error")
			}
		}

		return err
	}
}
