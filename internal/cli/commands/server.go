package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/mlhmz/dockermc-cloud-manager/internal/database"
	"github.com/mlhmz/dockermc-cloud-manager/internal/models"
	"github.com/mlhmz/dockermc-cloud-manager/internal/service"
	"github.com/spf13/cobra"
)

// Helper function to initialize services for server commands
func initializeServices() (*database.DB, *service.DockerService, *service.MinecraftServerService, func()) {
	// Initialize database
	db, err := database.New(cfg.DatabasePath, logger)
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}

	// Initialize Docker service
	dockerService, err := service.NewDockerService(logger)
	if err != nil {
		db.Close()
		logger.Error("Failed to initialize Docker service", "error", err)
		os.Exit(1)
	}

	// Initialize repositories
	serverRepo := database.NewServerRepository(db)

	// Initialize Minecraft server service
	mcService := service.NewMinecraftServerService(dockerService, serverRepo, logger)

	// Cleanup function
	cleanup := func() {
		dockerService.Close()
		db.Close()
	}

	return db, dockerService, mcService, cleanup
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage Minecraft servers",
	Long:  `Create, list, start, stop, and delete Minecraft servers.`,
}

var serverCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new Minecraft server",
	Long:  `Create a new Minecraft server with the specified name.`,
	Example: `  # Create a server with default settings
  dockermc-cloud-manager server create my-server

  # Create a server with custom settings
  dockermc-cloud-manager server create survival --max-players 50 --motd "Welcome!" --version 1.20.1`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		maxPlayers, _ := cmd.Flags().GetInt("max-players")
		motd, _ := cmd.Flags().GetString("motd")
		version, _ := cmd.Flags().GetString("version")

		ctx := context.Background()

		// Initialize services
		_, _, mcService, cleanup := initializeServices()
		defer cleanup()

		// Create server
		req := &models.CreateServerRequest{
			Name:       name,
			MaxPlayers: maxPlayers,
			MOTD:       motd,
			Version:    version,
		}

		logger.Info("Creating server", "name", name)
		server, err := mcService.CreateServer(ctx, req)
		if err != nil {
			logger.Error("Failed to create server", "error", err)
			os.Exit(1)
		}

		logger.Info("Server created successfully", "id", server.ID, "name", server.Name)

		// Print server details
		fmt.Printf("\n✓ Server created successfully!\n\n")
		fmt.Printf("ID:           %s\n", server.ID)
		fmt.Printf("Name:         %s\n", server.Name)
		fmt.Printf("Status:       %s\n", server.Status)
		fmt.Printf("Max Players:  %d\n", server.MaxPlayers)
		fmt.Printf("MOTD:         %s\n", server.MOTD)
		fmt.Printf("Container ID: %s\n", server.ContainerID)
		fmt.Printf("Volume ID:    %s\n", server.VolumeID)
		fmt.Printf("\nUse 'dockermc-cloud-manager server start %s' to start the server.\n", server.ID)
	},
}

var serverListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Minecraft servers",
	Long:  `List all Minecraft servers with their current status.`,
	Example: `  dockermc-cloud-manager server list
  dockermc-cloud-manager server list --output json`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		outputFormat, _ := cmd.Flags().GetString("output")

		// Initialize services
		_, _, mcService, cleanup := initializeServices()
		defer cleanup()

		// List servers
		servers, err := mcService.ListServers(ctx)
		if err != nil {
			logger.Error("Failed to list servers", "error", err)
			os.Exit(1)
		}

		if outputFormat == "json" {
			// JSON output
			data, _ := json.MarshalIndent(servers, "", "  ")
			fmt.Println(string(data))
			return
		}

		// Table output
		if len(servers) == 0 {
			fmt.Println("No servers found.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSTATUS\tMAX PLAYERS\tCREATED")
		for _, server := range servers {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
				server.ID[:8]+"...",
				server.Name,
				server.Status,
				server.MaxPlayers,
				server.CreatedAt.Format("2006-01-02 15:04"),
			)
		}
		w.Flush()
	},
}

var serverStartCmd = &cobra.Command{
	Use:     "start <server-id>",
	Short:   "Start a Minecraft server",
	Long:    `Start a stopped Minecraft server by its ID.`,
	Example: `  dockermc-cloud-manager server start abc123...`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverID := args[0]
		ctx := context.Background()

		// Initialize services
		_, _, mcService, cleanup := initializeServices()
		defer cleanup()

		// Start server
		logger.Info("Starting server", "id", serverID)
		if err := mcService.StartServer(ctx, serverID); err != nil {
			logger.Error("Failed to start server", "error", err)
			os.Exit(1)
		}

		logger.Info("Server started successfully", "id", serverID)
		fmt.Printf("✓ Server %s started successfully!\n", serverID)
	},
}

var serverStopCmd = &cobra.Command{
	Use:     "stop <server-id>",
	Short:   "Stop a Minecraft server",
	Long:    `Stop a running Minecraft server by its ID.`,
	Example: `  dockermc-cloud-manager server stop abc123...`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverID := args[0]
		ctx := context.Background()

		// Initialize services
		_, _, mcService, cleanup := initializeServices()
		defer cleanup()

		// Stop server
		logger.Info("Stopping server", "id", serverID)
		if err := mcService.StopServer(ctx, serverID); err != nil {
			logger.Error("Failed to stop server", "error", err)
			os.Exit(1)
		}

		logger.Info("Server stopped successfully", "id", serverID)
		fmt.Printf("✓ Server %s stopped successfully!\n", serverID)
	},
}

var serverDeleteCmd = &cobra.Command{
	Use:   "delete <server-id>",
	Short: "Delete a Minecraft server",
	Long:  `Delete a Minecraft server and all its data by its ID.`,
	Example: `  dockermc-cloud-manager server delete abc123...
  dockermc-cloud-manager server delete abc123... --force`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverID := args[0]
		force, _ := cmd.Flags().GetBool("force")
		ctx := context.Background()

		// Confirmation prompt
		if !force {
			fmt.Printf("⚠ Are you sure you want to delete server %s? This will remove all data. [y/N]: ", serverID)
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Cancelled.")
				return
			}
		}

		// Initialize services
		_, _, mcService, cleanup := initializeServices()
		defer cleanup()

		// Delete server
		logger.Info("Deleting server", "id", serverID)
		if err := mcService.DeleteServer(ctx, serverID); err != nil {
			logger.Error("Failed to delete server", "error", err)
			os.Exit(1)
		}

		logger.Info("Server deleted successfully", "id", serverID)
		fmt.Printf("✓ Server %s deleted successfully!\n", serverID)
	},
}

var serverInfoCmd = &cobra.Command{
	Use:   "info <server-id>",
	Short: "Show detailed information about a server",
	Long:  `Display detailed information about a specific Minecraft server.`,
	Example: `  dockermc-cloud-manager server info abc123...
  dockermc-cloud-manager server info abc123... --output json`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverID := args[0]
		ctx := context.Background()
		outputFormat, _ := cmd.Flags().GetString("output")

		// Initialize services
		_, _, mcService, cleanup := initializeServices()
		defer cleanup()

		// Get server info
		server, err := mcService.GetServer(ctx, serverID)
		if err != nil {
			logger.Error("Failed to get server info", "error", err)
			os.Exit(1)
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(server, "", "  ")
			fmt.Println(string(data))
			return
		}

		// Pretty print
		fmt.Printf("\nServer Information:\n")
		fmt.Printf("==================\n\n")
		fmt.Printf("ID:           %s\n", server.ID)
		fmt.Printf("Name:         %s\n", server.Name)
		fmt.Printf("Status:       %s\n", server.Status)
		fmt.Printf("Max Players:  %d\n", server.MaxPlayers)
		fmt.Printf("MOTD:         %s\n", server.MOTD)
		fmt.Printf("Container ID: %s\n", server.ContainerID)
		fmt.Printf("Volume ID:    %s\n", server.VolumeID)
		fmt.Printf("Created:      %s\n", server.CreatedAt.Format(time.RFC1123))
		fmt.Printf("Updated:      %s\n", server.UpdatedAt.Format(time.RFC1123))
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Create command
	serverCmd.AddCommand(serverCreateCmd)
	serverCreateCmd.Flags().IntP("max-players", "m", 20, "Maximum number of players")
	serverCreateCmd.Flags().StringP("motd", "d", "", "Message of the day")
	serverCreateCmd.Flags().StringP("version", "v", "LATEST", "Minecraft version")

	// List command
	serverCmd.AddCommand(serverListCmd)
	serverListCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")

	// Start command
	serverCmd.AddCommand(serverStartCmd)

	// Stop command
	serverCmd.AddCommand(serverStopCmd)

	// Delete command
	serverCmd.AddCommand(serverDeleteCmd)
	serverDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	// Info command
	serverCmd.AddCommand(serverInfoCmd)
	serverInfoCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
}
