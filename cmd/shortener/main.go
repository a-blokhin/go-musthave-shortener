package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-musthave-shortener/internal/config"
	"go-musthave-shortener/internal/di/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.ParseFlags()
	diConfig := app.NewConfig(cfg.ServerAddress, cfg.BaseURL)

	di := app.DI{}
	di.Init(diConfig)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)

	go func() {
		log.Printf("Starting server on %s", diConfig.ServerAddress)
		if err := di.StartServer(); err != nil {
			errCh <- fmt.Errorf("server error: %w", err)
		}
	}()

	select {
	case <-stop:
		log.Println("Shutting down server gracefully...")
	case err := <-errCh:
		log.Printf("Error: %v", err)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancel()

	if err := di.StopServer(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server stopped successfully")
}
