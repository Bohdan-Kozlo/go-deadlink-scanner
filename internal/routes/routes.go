package routes

import (
	"go-deadlink-scanner/internal/scanner"
	"go-deadlink-scanner/internal/user"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App, userHandler *user.Handler,
	scannerHandler *scanner.Handler) {
	app.Get("/register", userHandler.RegisterPage)
	app.Get("/login", userHandler.LoginPage)

	userGroup := app.Group("/api/user")
	userGroup.Post("/register", userHandler.Register)
	userGroup.Post("/login", userHandler.Login)

	scannerGroup := app.Group("/api/scanner")
	_ = scannerGroup
}
