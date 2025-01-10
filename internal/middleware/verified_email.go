package middleware

// import (
// 	"codebase-app/internal/repositories/pg"

// 	"github.com/gofiber/fiber/v2"
// 	"github.com/rs/zerolog/log"
// )

// func VerifiedEmailMiddleware(c *fiber.Ctx) error {
// 	var (
// 		ctx = c.Context()
// 	)

// 	userId, ok := c.Locals("user_id").(int64)
// 	if !ok {
// 		log.Error().Msg("middleware::VerifiedEmailMiddleware - Unauthorized [User Id not set]")
// 		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
// 			"message": "Your email is not verified. Please verify your email before inquire the product.",
// 			"success": false,
// 		})
// 	}

// 	authRepo := pg.NewAuthRepository()

// 	user, err := authRepo.FindById(ctx, userId)
// 	if err != nil {
// 		log.Error().Err(err).Msg("middleware::VerifiedEmailMiddleware - Unauthorized [User not found]")
// 		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
// 			"message": "Your email is not verified. Please verify your email before inquire the product.",
// 			"success": false,
// 		})
// 	}

// 	if user.EmailVerifiedAt == nil {
// 		log.Error().Msg("middleware::VerifiedEmailMiddleware - Unauthorized [Email not verified]")
// 		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
// 			"message": "Your email is not verified. Please verify your email before inquire the product.",
// 			"success": false,
// 		})
// 	}

// 	return c.Next()
// }
