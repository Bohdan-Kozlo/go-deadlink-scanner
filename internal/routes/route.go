package routes

import (
	"go-deadlink-scanner/internal/auth"
	"go-deadlink-scanner/internal/scanner"
	"go-deadlink-scanner/internal/user"

	"github.com/gofiber/fiber/v2"
)

type Router struct {
	app            *fiber.App
	userHandler    *user.Handler
	scannerHandler *scanner.Handler
	authMiddleware *auth.Middleware
}

func New(app *fiber.App, uh *user.Handler, sh *scanner.Handler, am *auth.Middleware) *Router {
	return &Router{app: app, userHandler: uh, scannerHandler: sh, authMiddleware: am}
}

func (r *Router) Register() {
	r.app.Get("/register", r.userHandler.RegisterPage)
	r.app.Get("/login", r.userHandler.LoginPage)

	r.app.Get("/scan", r.authMiddleware.RequireAuth(), r.scannerHandler.ScanPage)

	r.app.Post("/logout", r.userHandler.Logout)

	userGroup := r.app.Group("/api/user")
	userGroup.Post("/register", r.userHandler.Register)
	userGroup.Post("/login", r.userHandler.Login)
	userGroup.Post("/logout", r.authMiddleware.RequireAuth(), r.userHandler.Logout)

	scannerGroup := r.app.Group("/api/scanner", r.authMiddleware.RequireAuth())
	scannerGroup.Post("/scan", r.scannerHandler.StartScan)
}
