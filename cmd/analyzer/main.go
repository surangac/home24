package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"home24/internal/config"
	"home24/internal/handlers"
	"home24/pkg/logger"
)

func main() {
	log := logger.New()
	fmt.Println("Web Page Analyzer - Starting...")
	log.Info("starting web page analyzer application")

	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/application.yaml"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Error("failed to load configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Create server
	router := handlers.NewRouter(log)
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  parseDuration(cfg.Server.ReadTimeout),
		WriteTimeout: parseDuration(cfg.Server.WriteTimeout),
		IdleTimeout:  parseDuration(cfg.Server.IdleTimeout),
	}

	fmt.Printf("Server starting on port %s...\n", cfg.Server.Port)
	log.Info("server starting", slog.String("port", cfg.Server.Port))
	// Error channel for server errors
	serverErrors := make(chan error, 1)

	go func() {
		log.Info("starting server", slog.String("port", cfg.Server.Port))
		err := srv.ListenAndServe()
		if err != nil {
			serverErrors <- err
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal or server error
	select {
	case err := <-serverErrors:
		log.Error("server error", slog.String("error", err.Error()))
		os.Exit(1)
	case <-quit:
		log.Info("shutting down server...")

		// Create context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Try to gracefully shutdown the server
		err := srv.Shutdown(ctx)
		if err != nil {
			log.Error("server forced to shutdown", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}

	log.Info("server exited properly")
	fmt.Println("Server shutdown complete")
}

// parseDuration parses a duration string (e.g., "10s", "30s", "120s")
func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		// Return a default duration if parsing fails
		return 10 * time.Second
	}
	return d
}
