package app

import (
	"context"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

const shutdownTimeout = 10 * time.Second

type App struct {
	fiber *fiber.App
	port  string
}

func New() *App {
	f := fiber.New(fiber.Config{
		BodyLimit: 12 * 1024 * 1024, // uploads are up to 10 MB
	})
	f.Use(logger.New(), recover.New())

	api := f.Group("/api/v1")
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"status": "ok"}})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &App{fiber: f, port: port}
}

// Run serves until ctx is cancelled, then shuts down gracefully.
func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() { errCh <- a.fiber.Listen(":" + a.port) }()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return a.fiber.ShutdownWithTimeout(shutdownTimeout)
	}
}
