package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/api/dto"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/games/d2"
)

// AdminMiddleware checks if the authenticated user is an admin
func AdminMiddleware(repo *d2.Repository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := GetUserID(c)
		if userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
				Error:   "unauthorized",
				Message: "Authentication required",
				Code:    401,
			})
		}

		isAdmin, err := repo.IsAdmin(c.Context(), userID)
		if err != nil || !isAdmin {
			return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
				Error:   "forbidden",
				Message: "Admin access required",
				Code:    403,
			})
		}

		return c.Next()
	}
}
