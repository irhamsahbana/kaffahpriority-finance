package middleware

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
)

func RequestID(c *fiber.Ctx) error {
	requestID := c.Get("X-Request-ID")
	if requestID == "" {
		requestID = ulid.Make().String()
	}
	c.Set("X-Request-ID", requestID)
	c.Context().SetUserValue("request_id", requestID)

	return c.Next()
}

func WithAppLogger(base zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestId, _ := c.Context().UserValue("request_id").(string)
		subLogger := base.With().Str("request_id", requestId).Logger()

		// inject ke context bawaan Go
		ctx := subLogger.WithContext(c.UserContext())
		// Also inject request_id into context for tracing reconstruction
		ctx = context.WithValue(ctx, "request_id", requestId)

		c.SetUserContext(ctx)

		return c.Next()
	}

}
