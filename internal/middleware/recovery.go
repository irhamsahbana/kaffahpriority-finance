package middleware

import (
	"codebase-app/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// Recover is a middleware that recovers from panics, logs the error,
// and returns a JSON response with status 500.
// It should be placed after Tracing middleware to ensure the panic is recorded in the span.
func Recover() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				var requestID string
				if v := c.UserContext().Value("request_id"); v != nil {
					requestID, _ = v.(string)
				}
				if requestID == "" {
					requestID = "unknown"
				}

				// This log will be intercepted by the Tracing middleware's logger
				// and recorded as an error event in the active span.
				log.Ctx(c.UserContext()).Error().
					Str("request_id", requestID).
					Interface("error", r).
					Msg("Recovered from panic")

				// Return JSON response
				c.Status(fiber.StatusInternalServerError).JSON(response.Error("internal server error"))
			}
		}()
		return c.Next()
	}
}
