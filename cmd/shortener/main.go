package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-musthave-shortener/internal/di/app"
)

const (
	serverAddress = "localhost:8080"
	baseURL       = "http://localhost:8080"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := app.NewConfig(serverAddress, baseURL)

	di := app.DI{}
	di.Init(config)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)

	go func() {
		log.Printf("Starting server on %s", config.ServerAddress)
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
