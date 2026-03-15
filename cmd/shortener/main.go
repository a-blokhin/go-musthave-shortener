package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"go-musthave-shortener/internal/config"
	"go-musthave-shortener/internal/di/app"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	printBuildInfo()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	cfg := config.ParseConfig()

	di := app.DI{}
	if err := di.Init(cfg); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	errCh := make(chan error, 1)

	go func() {
		log.Printf("Starting server on %s", cfg.ServerAddress)
		if err := di.StartServer(); err != nil {
			errCh <- fmt.Errorf("server error: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("Shutting down server gracefully...")
	case err := <-errCh:
		log.Printf("Error: %v", err)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := di.StopServer(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server stopped successfully")
}

func printBuildInfo() {
	version := buildVersion
	if version == "" {
		version = "N/A"
	}
	date := buildDate
	if date == "" {
		date = "N/A"
	}
	commit := buildCommit
	if commit == "" {
		commit = "N/A"
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" && commit == "N/A" {
				commit = setting.Value
			}
			if setting.Key == "vcs.time" && date == "N/A" {
				date = setting.Value
			}
		}
	}

	fmt.Printf("Build version: %s\n", version)
	fmt.Printf("Build date: %s\n", date)
	fmt.Printf("Build commit: %s\n", commit)
}
