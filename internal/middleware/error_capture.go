package middleware

import (
	"codebase-app/internal/infrastructure/errtrack"
	"errors"

	"github.com/gofiber/fiber/v2"
)

// ErrorCapture reports requests that end with a 5xx response to error tracking.
// It must run after RequestID (it keys buffered error logs by request ID) and
// before Recover (so a recovered panic is seen as a 500 that was already reported).
func ErrorCapture() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID, _ := c.Context().UserValue("request_id").(string)
		errtrack.BeginRequest(requestID)

		err := c.Next()

		// A returned error is turned into a response by Fiber's error handler later,
		// so derive the status the same way it will.
		status := c.Response().StatusCode()
		if err != nil {
			status = fiber.StatusInternalServerError
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				status = fiberErr.Code
			}
		}

		userID, _ := c.Locals("user_id").(string)
		errtrack.EndRequest(errtrack.RequestInfo{
			RequestID: requestID,
			Method:    c.Method(),
			Route:     c.Route().Path,
			Path:      c.Path(),
			Status:    status,
			UserID:    userID,
			Err:       err,
		})

		return err
	}
}
