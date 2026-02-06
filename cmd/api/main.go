package main

import (
	"errors"
	"log"
	"net/http"

	"news-go/internal/app"
	"news-go/internal/config"
)

func main() {
	cfg := config.Load()
	if err := app.Run(cfg); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}
