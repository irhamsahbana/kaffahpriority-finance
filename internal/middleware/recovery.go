package middleware

import (
	"codebase-app/internal/infrastructure/errtrack"
	"codebase-app/pkg/response"
	"runtime/debug"

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
				// SkipField: the panic is reported below with a better stack trace.
				log.Ctx(c.UserContext()).Error().
					Bool(errtrack.SkipField, true).
					Str("request_id", requestID).
					Interface("error", r).
					Str("stack", string(debug.Stack())).
					Msg("Recovered from panic")

				// Captured inside the deferred call so the stack still includes the panic site.
				errtrack.CaptureRecovered(c.UserContext(), r, map[string]string{
					"http.method": c.Method(),
					"http.route":  c.Route().Path,
				})
				errtrack.MarkReported(requestID)

				// Return JSON response
				c.Status(fiber.StatusInternalServerError).JSON(response.Error("internal server error"))
			}
		}()
		return c.Next()
	}
}
