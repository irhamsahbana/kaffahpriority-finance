package route

import (
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	m "codebase-app/internal/middleware"
	userHandler "codebase-app/internal/module/user/handler"

	"codebase-app/pkg/response"
)

func SetupRoutes(app *fiber.App) {
	app.Get("/storage/private/:filename", m.ValidateSignedURL, storageFile)

	userHandler.NewUserHandler().Register(app.Group("/users"))

	// // fallback route
	app.Use(func(c *fiber.Ctx) error {
		var (
			method = c.Method()                       // get the request method
			path   = c.Path()                         // get the request path
			query  = c.Context().QueryArgs().String() // get all query params
			ua     = c.Get("User-Agent")              // get the request user agent
			ip     = c.IP()                           // get the request IP
		)

		log.Debug().
			Str("method", method).
			Str("path", path).
			Str("query", query).
			Str("ua", ua).
			Str("ip", ip).
			Msg("Route not found.")
		return c.Status(fiber.StatusNotFound).JSON(response.Error("Route not found."))
	})
}

func storageFile(c *fiber.Ctx) error {
	var (
		fileName = c.Params("filename")
		filePath = filepath.Join("storage", "private", fileName)
	)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Error().Err(err).Any("url", filePath).Msg("handler::getWAC - File not found")
		return c.Status(fiber.StatusNotFound).JSON(response.Error("File not found"))
	}

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		log.Error().Err(err).Any("url", filePath).Msg("handler::getWAC - Failed to read file")
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error(err.Error()))
	}

	return c.Send(fileBytes)
}
