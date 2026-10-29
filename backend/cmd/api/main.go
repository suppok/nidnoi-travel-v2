package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"backend/cmd/api/app"
)

// @title                      nidnoi travel API
// @version                    1.0
// @BasePath                   /api/v1
// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                "Bearer <access token>"
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	a, err := app.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if err := a.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
