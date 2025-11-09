package service

import (
	"context"
	"fmt"
	"io"

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
}

// NewMinecraftServerService creates a new Minecraft server service
func NewMinecraftServerService(dockerService *DockerService, repo *database.ServerRepository) *MinecraftServerService {
	return &MinecraftServerService{
		dockerService: dockerService,
		repo:          repo,
	}
}

// CreateServer creates a new Minecraft server
func (s *MinecraftServerService) CreateServer(ctx context.Context, req *models.CreateServerRequest) (*models.MinecraftServer, error) {
	// Generate unique ID
	serverID := uuid.New().String()

	// Create volume for persistent storage
	volumeName := fmt.Sprintf("mc-server-%s", serverID)
	vol, err := s.dockerService.client.VolumeCreate(ctx, volume.CreateOptions{
		Name: volumeName,
		Labels: map[string]string{
			"minecraft-server-id": serverID,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}

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
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}

	// Create container configuration
	containerConfig := &container.Config{
		Image: imageName,
		Env: []string{
			"EULA=TRUE",
			fmt.Sprintf("MAX_PLAYERS=%d", maxPlayers),
			fmt.Sprintf("MOTD=%s", motd),
			fmt.Sprintf("VERSION=%s", version),
			"TYPE=PAPER",
		},
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
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

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

	// Save to database
	if err := s.repo.Create(server); err != nil {
		// Cleanup container and volume on failure
		s.dockerService.client.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
		s.dockerService.client.VolumeRemove(ctx, vol.Name, true)
		return nil, fmt.Errorf("failed to save server to database: %w", err)
	}

	return server, nil
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
