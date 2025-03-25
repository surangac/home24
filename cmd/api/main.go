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

	"home24/internal/handlers"
	"home24/pkg/logger"
)

func main() {
	log := logger.New()
	fmt.Println("Web Page Analyzer - Starting...")
	log.Info("starting web page analyzer application")

	// Create server
	port := ":8080"
	router := handlers.NewRouter(log)
	srv := &http.Server{
		Addr:         port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	fmt.Println("Server starting on port 8080...")
	log.Info("server starting on port 8080")
	// Error channel for server errors
	serverErrors := make(chan error, 1)

	go func() {
		log.Info("starting server", slog.String("port", port))
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
