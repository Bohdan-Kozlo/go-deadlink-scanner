package auth

import (
	"time"

	db "go-deadlink-scanner/internal/database/sqlc"

	"github.com/gofiber/fiber/v2"
)

type Middleware struct {
	Queries *db.Queries
}

func NewMiddleware(queries *db.Queries) *Middleware {
	return &Middleware{
		Queries: queries,
	}
}

func (m *Middleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies("session_token")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing session token",
			})
		}

		sessionDb, err := m.Queries.GetSessionByToken(c.Context(), token)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Invalid session",
			})
		}

		if sessionDb.ExpiresAt.Before(time.Now()) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Session expired",
			})
		}

		c.Locals("user_id", sessionDb.UserID)

		return c.Next()
	}
}
