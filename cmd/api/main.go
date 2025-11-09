package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mlhmz/dockermc-cloud-manager/internal/api/routes"
	"github.com/mlhmz/dockermc-cloud-manager/internal/config"
	"github.com/mlhmz/dockermc-cloud-manager/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Docker service
	dockerService, err := service.NewDockerService()
	if err != nil {
		log.Fatalf("Failed to initialize Docker service: %v", err)
	}
	defer dockerService.Close()

	// Initialize Minecraft server service
	mcService := service.NewMinecraftServerService(dockerService)

	// Setup router
	router := routes.NewRouter(mcService)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the server
	go func() {
		log.Printf("Starting API server on port %d", cfg.Port)
		serverErrors <- srv.ListenAndServe()
	}()

	// Channel to listen for interrupt signal to terminate server gracefully
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or error
	select {
	case err := <-serverErrors:
		log.Fatalf("Error starting server: %v", err)

	case sig := <-shutdown:
		log.Printf("Received shutdown signal: %v. Starting graceful shutdown...", sig)

		// Give outstanding requests a deadline for completion
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Error during shutdown: %v", err)
			if err := srv.Close(); err != nil {
				log.Fatalf("Could not stop server gracefully: %v", err)
			}
		}

		log.Println("Server stopped gracefully")
	}
}
