package main

import (
	"database/sql"
	"go-deadlink-scanner/internal/config"
	"log"

	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.LoadConfig()

	connect, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer connect.Close()

	// queries := db.New(connect)

	app := fiber.New(fiber.Config{
		AppName: "Go Dead Link Scanner",
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	log.Printf("Server started on :%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
