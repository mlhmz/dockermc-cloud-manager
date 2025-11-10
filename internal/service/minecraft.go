package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/volume"
	"github.com/google/uuid"
	"github.com/mlhmz/dockermc-cloud-manager/internal/database"
	"github.com/mlhmz/dockermc-cloud-manager/internal/models"
)

// MinecraftServerService manages Minecraft server lifecycle
type MinecraftServerService struct {
	dockerService *DockerService
	repo          *database.ServerRepository
	proxyService  *ProxyService
	logger        *slog.Logger
}

// NewMinecraftServerService creates a new Minecraft server service
func NewMinecraftServerService(dockerService *DockerService, repo *database.ServerRepository, logger *slog.Logger) *MinecraftServerService {
	return &MinecraftServerService{
		dockerService: dockerService,
		repo:          repo,
		logger:        logger,
	}
}

// SetProxyService sets the proxy service (called after ProxyService is created to avoid circular dependency)
func (s *MinecraftServerService) SetProxyService(proxyService *ProxyService) {
	s.proxyService = proxyService
}

// CreateServer creates a new Minecraft server
func (s *MinecraftServerService) CreateServer(ctx context.Context, req *models.CreateServerRequest) (*models.MinecraftServer, error) {
	// Generate unique ID
	serverID := uuid.New().String()

	s.logger.InfoContext(ctx, "Creating new Minecraft server",
		"server_id", serverID,
		"server_name", req.Name,
		"max_players", req.MaxPlayers,
		"version", req.Version)

	// Create volume for persistent storage
	volumeName := fmt.Sprintf("mc-server-%s", serverID)
	vol, err := s.dockerService.client.VolumeCreate(ctx, volume.CreateOptions{
		Name: volumeName,
		Labels: map[string]string{
			"minecraft-server-id": serverID,
		},
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to create Docker volume",
			"server_id", serverID,
			"server_name", req.Name,
			"volume_name", volumeName,
			"error", err)
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}

	s.logger.DebugContext(ctx, "Created Docker volume",
		"server_id", serverID,
		"volume_name", volumeName)

	// Set defaults
	maxPlayers := req.MaxPlayers
	if maxPlayers == 0 {
		maxPlayers = 20
	}

	motd := req.MOTD
	if motd == "" {
		motd = fmt.Sprintf("Minecraft Server - %s", req.Name)
	}

	version := req.Version
	if version == "" {
		version = "LATEST"
	}

	// Pull the image if it doesn't exist
	imageName := "itzg/minecraft-server:latest"
	if err := s.dockerService.PullImage(ctx, imageName); err != nil {
		// Cleanup volume on failure
		s.dockerService.client.VolumeRemove(ctx, vol.Name, true)
		s.logger.ErrorContext(ctx, "Failed to pull Docker image",
			"server_id", serverID,
			"server_name", req.Name,
			"image", imageName,
			"error", err)
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}

	// Check if proxy exists to determine if we should configure for proxy mode
	hasProxy := false
	if s.proxyService != nil {
		if _, err := s.proxyService.EnsureProxyExists(ctx); err == nil {
			hasProxy = true
			s.logger.InfoContext(ctx, "Configuring server for proxy mode",
				"server_id", serverID)
		}
	}

	// Create container configuration
	env := []string{
		"EULA=TRUE",
		fmt.Sprintf("MAX_PLAYERS=%d", maxPlayers),
		fmt.Sprintf("MOTD=%s", motd),
		fmt.Sprintf("VERSION=%s", version),
		"TYPE=PAPER",
	}

	// Configure for legacy BungeeCord/Velocity forwarding if proxy exists
	if hasProxy {
		env = append(env,
			"ONLINE_MODE=FALSE",               // Must be false when behind proxy
			"PATCH_DEFINITIONS=/data/patches", // Directory containing patch definitions in volume
		)
	}

	containerConfig := &container.Config{
		Image: imageName,
		Env:   env,
		Labels: map[string]string{
			"minecraft-server-id":   serverID,
			"minecraft-server-name": req.Name,
		},
	}

	hostConfig := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/data", vol.Name),
		},
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	// Create container
	s.logger.DebugContext(ctx, "Creating Docker container",
		"server_id", serverID,
		"container_name", fmt.Sprintf("mc-server-%s", serverID))

	resp, err := s.dockerService.client.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		nil,
		nil,
		fmt.Sprintf("mc-server-%s", serverID),
	)
	if err != nil {
		// Cleanup volume on failure
		s.dockerService.client.VolumeRemove(ctx, vol.Name, true)
		s.logger.ErrorContext(ctx, "Failed to create Docker container",
			"server_id", serverID,
			"server_name", req.Name,
			"error", err)
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	s.logger.DebugContext(ctx, "Created Docker container",
		"server_id", serverID,
		"container_id", resp.ID)

	// Create server model
	server := &models.MinecraftServer{
		ID:          serverID,
		Name:        req.Name,
		ContainerID: resp.ID,
		VolumeID:    vol.Name,
		Status:      models.StatusCreating,
		MaxPlayers:  maxPlayers,
		MOTD:        motd,
	}

	// If configured for proxy, create the patch file in the volume BEFORE saving to database
	// This needs to happen before the container starts
	if hasProxy {
		s.logger.DebugContext(ctx, "Creating BungeeCord patch file for proxy compatibility",
			"server_id", serverID)

		if err := s.createBungeeCordPatchFileInVolume(ctx, vol.Name); err != nil {
			// Cleanup on failure
			s.dockerService.client.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
			s.dockerService.client.VolumeRemove(ctx, vol.Name, true)
			s.logger.ErrorContext(ctx, "Failed to create BungeeCord patch file",
				"server_id", serverID,
				"error", err)
			return nil, fmt.Errorf("failed to create patch file: %w", err)
		}
	}

	// Save to database
	s.logger.DebugContext(ctx, "Saving server to database",
		"server_id", serverID,
		"server_name", req.Name)

	if err := s.repo.Create(server); err != nil {
		// Cleanup container and volume on failure
		s.dockerService.client.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
		s.dockerService.client.VolumeRemove(ctx, vol.Name, true)
		s.logger.ErrorContext(ctx, "Failed to save server to database",
			"server_id", serverID,
			"server_name", req.Name,
			"error", err)
		return nil, fmt.Errorf("failed to save server to database: %w", err)
	}

	// Auto-connect server to proxy if proxy service is available
	if s.proxyService != nil {
		// Ensure proxy exists
		if _, err := s.proxyService.EnsureProxyExists(ctx); err != nil {
			// Log error but don't fail server creation
			// Server can still function standalone
			s.logger.WarnContext(ctx, "Failed to connect server to proxy",
				"server_id", serverID,
				"error", err)
		} else {
			// Connect server to proxy network
			if err := s.proxyService.ConnectServerToProxy(ctx, server); err != nil {
				// Log error but don't fail server creation
				s.logger.WarnContext(ctx, "Failed to connect server to proxy network",
					"server_id", serverID,
					"error", err)
			} else {
				s.logger.InfoContext(ctx, "Connected server to proxy network",
					"server_id", serverID)

				// Regenerate proxy config to include new server
				if err := s.proxyService.RegenerateProxyConfig(ctx); err != nil {
					s.logger.WarnContext(ctx, "Failed to regenerate proxy config",
						"server_id", serverID,
						"error", err)
				}
			}
		}
	}

	s.logger.InfoContext(ctx, "Server created successfully",
		"server_id", serverID,
		"server_name", req.Name,
		"container_id", resp.ID)

	return server, nil
}

// createBungeeCordPatchFileInVolume creates a patch definition file in the volume using a temporary container
func (s *MinecraftServerService) createBungeeCordPatchFileInVolume(ctx context.Context, volumeName string) error {
	// Pull alpine image if not present
	alpineImage := "alpine:latest"
	if err := s.dockerService.PullImage(ctx, alpineImage); err != nil {
		return fmt.Errorf("failed to pull alpine image: %w", err)
	}

	// Patch definition to enable bungeecord in spigot.yml
	// Note: Each file in the patches directory is a single patch, not an array
	patchContent := `{
  "file": "/data/spigot.yml",
  "ops": [
    {
      "$set": {
        "path": "$.settings.bungeecord",
        "value": true,
        "value-type": "bool"
      }
    }
  ]
}`

	// Use a temporary alpine container to write the file to the volume
	tempContainerConfig := &container.Config{
		Image: "alpine:latest",
		Cmd: []string{"sh", "-c", fmt.Sprintf(`
mkdir -p /data/patches && cat > /data/patches/bungeecord.json << 'PATCHEOF'
%s
PATCHEOF
`, patchContent)},
	}

	tempHostConfig := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/data", volumeName),
		},
	}

	// Create temporary container
	tempResp, err := s.dockerService.client.ContainerCreate(ctx, tempContainerConfig, tempHostConfig, nil, nil, "")
	if err != nil {
		return fmt.Errorf("failed to create temp container: %w", err)
	}
	defer s.dockerService.client.ContainerRemove(ctx, tempResp.ID, container.RemoveOptions{Force: true})

	// Start and wait for the temp container to finish
	if err := s.dockerService.client.ContainerStart(ctx, tempResp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start temp container: %w", err)
	}

	// Wait for container to finish
	statusCh, errCh := s.dockerService.client.ContainerWait(ctx, tempResp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("error waiting for temp container: %w", err)
		}
	case <-statusCh:
	}

	return nil
}

// ListServers returns all servers
func (s *MinecraftServerService) ListServers(ctx context.Context) ([]*models.MinecraftServer, error) {
	return s.repo.FindAll()
}

// GetServer returns a specific server by ID
func (s *MinecraftServerService) GetServer(ctx context.Context, id string) (*models.MinecraftServer, error) {
	return s.repo.FindByID(id)
}

// StartServer starts a Minecraft server
func (s *MinecraftServerService) StartServer(ctx context.Context, id string) error {
	server, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if err := s.dockerService.client.ContainerStart(ctx, server.ContainerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	server.Status = models.StatusRunning
	return s.repo.Update(server)
}

// StopServer stops a Minecraft server
func (s *MinecraftServerService) StopServer(ctx context.Context, id string) error {
	server, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	timeout := 30
	if err := s.dockerService.client.ContainerStop(ctx, server.ContainerID, container.StopOptions{
		Timeout: &timeout,
	}); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	server.Status = models.StatusStopped
	return s.repo.Update(server)
}

// DeleteServer removes a Minecraft server and its resources
func (s *MinecraftServerService) DeleteServer(ctx context.Context, id string) error {
	server, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	// Stop container if running
	timeout := 30
	s.dockerService.client.ContainerStop(ctx, server.ContainerID, container.StopOptions{
		Timeout: &timeout,
	})

	// Remove container
	if err := s.dockerService.client.ContainerRemove(ctx, server.ContainerID, container.RemoveOptions{
		Force: true,
	}); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	// Remove volume
	if err := s.dockerService.client.VolumeRemove(ctx, server.VolumeID, true); err != nil {
		return fmt.Errorf("failed to remove volume: %w", err)
	}

	// Remove from database
	return s.repo.Delete(id)
}

// GetServerLogs retrieves logs from a server's Docker container
func (s *MinecraftServerService) GetServerLogs(ctx context.Context, containerID string, follow bool, tail string) (io.ReadCloser, error) {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Tail:       tail,
		Timestamps: false,
	}

	logs, err := s.dockerService.client.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}

	return logs, nil
}

// ExecuteCommand executes a Minecraft command via rcon-cli in the container
func (s *MinecraftServerService) ExecuteCommand(ctx context.Context, containerID string, command string) (string, error) {
	// Create exec configuration to run rcon-cli
	execConfig := container.ExecOptions{
		Cmd:          []string{"rcon-cli", command},
		AttachStdout: true,
		AttachStderr: true,
	}

	// Create the exec instance
	execResp, err := s.dockerService.client.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create exec: %w", err)
	}

	// Attach to the exec instance
	attachResp, err := s.dockerService.client.ContainerExecAttach(ctx, execResp.ID, container.ExecStartOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to attach to exec: %w", err)
	}
	defer attachResp.Close()

	// Read the output
	output, err := io.ReadAll(attachResp.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to read exec output: %w", err)
	}

	return string(output), nil
}
