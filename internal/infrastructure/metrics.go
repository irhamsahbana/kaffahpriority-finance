package infrastructure

import (
	"codebase-app/internal/infrastructure/config"
	"encoding/base64"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

func InitializeMetrics(app *fiber.App) {
	handler := fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())

	app.Get("/metrics", func(c *fiber.Ctx) error {
		u := config.Envs.Guard.Metrics.BasicAuthUser
		p := config.Envs.Guard.Metrics.BasicAuthPass
		if u != "" && p != "" {
			auth := c.Get("Authorization")
			if !strings.HasPrefix(auth, "Basic ") {
				return c.SendStatus(fiber.StatusUnauthorized)
			}
			enc := strings.TrimPrefix(auth, "Basic ")
			dec, err := base64.StdEncoding.DecodeString(enc)
			if err != nil {
				return c.SendStatus(fiber.StatusUnauthorized)
			}
			parts := strings.SplitN(string(dec), ":", 2)
			if len(parts) != 2 {
				return c.SendStatus(fiber.StatusUnauthorized)
			}
			if parts[0] != u || parts[1] != p {
				return c.SendStatus(fiber.StatusUnauthorized)
			}
		}
		handler(c.Context())
		return nil
	})
}
