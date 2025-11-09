package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/volume"
	"github.com/google/uuid"
	"github.com/mlhmz/dockermc-cloud-manager/internal/models"
)

// MinecraftServerService manages Minecraft server lifecycle
type MinecraftServerService struct {
	dockerService *DockerService
	servers       map[string]*models.MinecraftServer
	mu            sync.RWMutex
}

// NewMinecraftServerService creates a new Minecraft server service
func NewMinecraftServerService(dockerService *DockerService) *MinecraftServerService {
	return &MinecraftServerService{
		dockerService: dockerService,
		servers:       make(map[string]*models.MinecraftServer),
	}
}

// CreateServer creates a new Minecraft server
func (s *MinecraftServerService) CreateServer(ctx context.Context, req *models.CreateServerRequest) (*models.MinecraftServer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

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
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Store server
	s.servers[serverID] = server

	return server, nil
}

// ListServers returns all servers
func (s *MinecraftServerService) ListServers(ctx context.Context) ([]*models.MinecraftServer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	servers := make([]*models.MinecraftServer, 0, len(s.servers))
	for _, server := range s.servers {
		servers = append(servers, server)
	}

	return servers, nil
}

// GetServer returns a specific server by ID
func (s *MinecraftServerService) GetServer(ctx context.Context, id string) (*models.MinecraftServer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	server, exists := s.servers[id]
	if !exists {
		return nil, fmt.Errorf("server not found")
	}

	return server, nil
}

// StartServer starts a Minecraft server
func (s *MinecraftServerService) StartServer(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	server, exists := s.servers[id]
	if !exists {
		return fmt.Errorf("server not found")
	}

	if err := s.dockerService.client.ContainerStart(ctx, server.ContainerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	server.Status = models.StatusRunning
	server.UpdatedAt = time.Now()

	return nil
}

// StopServer stops a Minecraft server
func (s *MinecraftServerService) StopServer(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	server, exists := s.servers[id]
	if !exists {
		return fmt.Errorf("server not found")
	}

	timeout := 30
	if err := s.dockerService.client.ContainerStop(ctx, server.ContainerID, container.StopOptions{
		Timeout: &timeout,
	}); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	server.Status = models.StatusStopped
	server.UpdatedAt = time.Now()

	return nil
}

// DeleteServer removes a Minecraft server and its resources
func (s *MinecraftServerService) DeleteServer(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	server, exists := s.servers[id]
	if !exists {
		return fmt.Errorf("server not found")
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

	// Remove from map
	delete(s.servers, id)

	return nil
}
