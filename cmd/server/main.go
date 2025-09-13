package main

import (
	"database/sql"
	"go-deadlink-scanner/internal/config"
	db "go-deadlink-scanner/internal/database/sqlc"
	"go-deadlink-scanner/internal/routes"
	"go-deadlink-scanner/internal/scanner"
	"go-deadlink-scanner/internal/user"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
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

	store := session.New(session.Config{
		CookieHTTPOnly: true,
		Expiration:     cfg.SessionMaxAge,
		KeyLookup:      "cookie:session_id",
		CookieSecure:   false,
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	userService := user.NewService(queries, store)
	scannerService := scanner.NewService(queries)

	userHandler := user.NewHandler(userService)
	scannerHandler := scanner.NewHandler(scannerService)

	routes.Setup(app, userHandler, scannerHandler)

	log.Printf("Server started on :%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
