package middleware

import (
	"codebase-app/internal/infrastructure/metrics"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Prometheus() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var (
			start = time.Now()
			err   = c.Next()
		)

		if err != nil {
			return err
		}

		var (
			duration = time.Since(start).Seconds()
			method   = c.Method()
			path     = c.Route().Path
			status   = strconv.Itoa(c.Response().StatusCode())
		)

		metrics.HttpRequestsTotal.
			WithLabelValues(method, path, status).
			Inc()

		metrics.HttpRequestDuration.
			WithLabelValues(method, path).
			Observe(duration)

		return err
	}
}
