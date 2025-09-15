package main

import (
	"database/sql"
	"go-deadlink-scanner/internal/auth"
	"go-deadlink-scanner/internal/config"
	"go-deadlink-scanner/internal/database/sqlc" // added
	"go-deadlink-scanner/internal/routes"
	"go-deadlink-scanner/internal/scanner"
	"go-deadlink-scanner/internal/user"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.LoadConfig()

	connect, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer connect.Close()

	queries := db.New(connect)

	app := fiber.New(fiber.Config{
		AppName: "Go Dead Link Scanner",
	})

	app.Use(logger.New())

	app.Static("/static", "./static")

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	userService := user.NewService(queries)
	scannerService := scanner.NewService(queries)

	userHandler := user.NewHandler(userService)
	scannerHandler := scanner.NewHandler(scannerService)

	middleware := auth.NewMiddleware(queries)

	r := routes.New(app, userHandler, scannerHandler, middleware)
	r.Register()

	log.Printf("Server started on :%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
