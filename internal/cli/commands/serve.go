package commands

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/mlhmz/dockermc-cloud-manager/internal/api/routes"
	"github.com/mlhmz/dockermc-cloud-manager/internal/database"
	"github.com/mlhmz/dockermc-cloud-manager/internal/service"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve [port]",
	Short: "Start the REST API server",
	Long: `Start the REST API server for managing Minecraft servers.

The server provides a REST API with Swagger documentation at /swagger/.
If no port is specified, it uses the API_PORT environment variable or defaults to 8080.`,
	Example: `  # Start on default port (8080 or API_PORT)
  dockermc-cloud-manager serve

  # Start on specific port
  dockermc-cloud-manager serve 9000`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Determine port
		port := cfg.Port
		if len(args) > 0 {
			parsedPort, err := strconv.Atoi(args[0])
			if err != nil {
				logger.Error("Invalid port number", "port", args[0], "error", err)
				os.Exit(1)
			}
			port = parsedPort
		}

		logger.Info("Starting Docker Minecraft Cloud Manager API",
			"port", port,
			"docker_network", cfg.DockerNetwork,
			"minecraft_image", cfg.MinecraftImage,
			"database_path", cfg.DatabasePath,
		)

		// Initialize database
		db, err := database.New(cfg.DatabasePath, logger)
		if err != nil {
			logger.Error("Failed to initialize database", "error", err)
			os.Exit(1)
		}
		defer db.Close()

		// Initialize Docker service
		dockerService, err := service.NewDockerService(logger)
		if err != nil {
			logger.Error("Failed to initialize Docker service", "error", err)
			os.Exit(1)
		}
		defer dockerService.Close()

		logger.Info("Docker service initialized")

		// Initialize repositories
		serverRepo := database.NewServerRepository(db)

		// Initialize Minecraft server service
		mcService := service.NewMinecraftServerService(dockerService, serverRepo)

		// Setup router
		router := routes.NewRouter(mcService, logger)

		// Create HTTP server
		srv := &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		// Channel to listen for errors coming from the listener
		serverErrors := make(chan error, 1)

		// Start the server
		go func() {
			logger.Info("API server listening", "port", port, "address", srv.Addr)
			logger.Info("Swagger UI available", "url", fmt.Sprintf("http://localhost:%d/swagger/", port))
			serverErrors <- srv.ListenAndServe()
		}()

		// Channel to listen for interrupt signal to terminate server gracefully
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

		// Block until we receive a signal or error
		select {
		case err := <-serverErrors:
			logger.Error("Error starting server", "error", err)
			os.Exit(1)

		case sig := <-shutdown:
			logger.Info("Received shutdown signal, starting graceful shutdown", "signal", sig.String())

			// Give outstanding requests a deadline for completion
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			// Attempt graceful shutdown
			if err := srv.Shutdown(ctx); err != nil {
				logger.Error("Error during shutdown", "error", err)
				if err := srv.Close(); err != nil {
					logger.Error("Could not stop server gracefully", "error", err)
					os.Exit(1)
				}
			}

			logger.Info("Server stopped gracefully")
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
