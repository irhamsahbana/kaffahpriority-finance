package route

import (
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	m "codebase-app/internal/middleware"
	masterHandler "codebase-app/internal/module/master/handler"
	reportHandler "codebase-app/internal/module/report/handler"
	userHandler "codebase-app/internal/module/user/handler"

	"codebase-app/pkg/response"
)

func SetupRoutes(app *fiber.App) {
	app.Get("/storage/private/:filename", m.ValidateSignedURL, storageFile)

	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(response.Success(nil, "Pong"))
	})

	userHandler.NewUserHandler().Register(app.Group("/users"))
	reportHandler.NewReportHandler().Register(app.Group("/reports"))
	masterHandler.NewMasterHandler().Register(app.Group("/masters"))

	// db := adapter.Adapters.Postgres

	// workerLimit := 15

	// jobs := make(chan string, 1000)
	// results := make(chan string, 1000)
	// errors := make(chan error, 1000)

	// var wg sync.WaitGroup

	// for w := 1; w <= workerLimit; w++ {
	// 	wg.Add(1)
	// 	go worker(db, jobs, results, errors, &wg)
	// }

	// for i := 0; i < 1000; i++ {
	// 	jobs <- `
	// 		INSERT INTO students (id, identifier, name) VALUES (
	// 		'` + ulid.Make().String() + `',
	// 		'` + ulid.Make().String() + `',
	// 		'` + gofakeit.Name() + `'
	// 		)
	// 	`
	// }

	// close(jobs)

	// wg.Wait()
	// close(results)
	// close(errors)

	// for result := range results {
	// 	log.Info().Str("result", result).Msg("Success")
	// }
	// for err := range errors {
	// 	log.Error().Err(err).Msg("Error")
	// }

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

// func worker(db *sqlx.DB, jobs <-chan string, results chan<- string, errors chan<- error, wg *sync.WaitGroup) {
// 	defer wg.Done()

// 	for query := range jobs {
// 		_, err := db.Exec(query)
// 		if err != nil {
// 			errors <- fmt.Errorf("failed insert %s: %v", query, err)
// 			continue
// 		}

// 		results <- "success insert " + query
// 	}
// }
