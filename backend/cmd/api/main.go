package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"backend/cmd/api/app"
)

// @title       nidnoi travel API
// @version     1.0
// @BasePath    /api/v1
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.New().Run(ctx); err != nil {
		log.Fatal(err)
	}
}
