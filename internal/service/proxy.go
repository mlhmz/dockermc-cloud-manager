package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/go-connections/nat"
	"github.com/mlhmz/dockermc-cloud-manager/internal/database"
	"github.com/mlhmz/dockermc-cloud-manager/internal/models"
)

const (
	MinecraftNetworkName = "minecraft-network"
	VelocityImage        = "itzg/bungeecord:latest"
	DefaultProxyPort     = 25565
)

// ProxyService manages the single Velocity proxy server
type ProxyService struct {
	dockerService *DockerService
	proxyRepo     *database.ProxyRepository
	serverRepo    *database.ServerRepository
	logger        *slog.Logger
}

// NewProxyService creates a new proxy service
func NewProxyService(
	dockerService *DockerService,
	proxyRepo *database.ProxyRepository,
	serverRepo *database.ServerRepository,
	logger *slog.Logger,
) *ProxyService {
	return &ProxyService{
		dockerService: dockerService,
		proxyRepo:     proxyRepo,
		serverRepo:    serverRepo,
		logger:        logger,
	}
}

// EnsureProxyExists creates the proxy if it doesn't exist
func (s *ProxyService) EnsureProxyExists(ctx context.Context) (*models.ProxyServer, error) {
	s.logger.DebugContext(ctx, "Checking if proxy exists")

	// Check if proxy already exists
	proxy, err := s.proxyRepo.FindByID(models.SingleProxyID)
	if err == nil {
		s.logger.InfoContext(ctx, "Proxy already exists", "proxy_id", proxy.ID)
		return proxy, nil // Proxy exists
	}

	s.logger.InfoContext(ctx, "Proxy does not exist, creating new proxy")
	// Create the proxy
	return s.createProxy(ctx)
}

func (s *ProxyService) UpdateProxy(ctx context.Context, proxy *models.ProxyServer) (*models.ProxyServer, error) {
	s.logger.InfoContext(ctx, "Updating proxy configuration", "proxy_id", proxy.ID)

	// Update the proxy configuration
	if err := s.proxyRepo.Update(proxy); err != nil {
		s.logger.ErrorContext(ctx, "Failed to update proxy", "proxy_id", proxy.ID, "error", err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "Proxy updated successfully", "proxy_id", proxy.ID)
	return proxy, nil
}

// createProxy creates the single Velocity proxy server
func (s *ProxyService) createProxy(ctx context.Context) (*models.ProxyServer, error) {
	s.logger.InfoContext(ctx, "Creating proxy server")

	// Create volume for proxy configuration
	volumeName := "mc-proxy-main"
	s.logger.DebugContext(ctx, "Creating volume for proxy", "volume_name", volumeName)

	vol, err := s.dockerService.client.VolumeCreate(ctx, volume.CreateOptions{
		Name: volumeName,
		Labels: map[string]string{
			"minecraft-proxy": "main",
		},
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to create proxy volume", "volume_name", volumeName, "error", err)
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}

	s.logger.DebugContext(ctx, "Volume created successfully", "volume_name", vol.Name)

	// Pull the Velocity image
	s.logger.InfoContext(ctx, "Pulling Velocity image", "image", VelocityImage)
	if err := s.dockerService.PullImage(ctx, VelocityImage); err != nil {
		s.dockerService.client.VolumeRemove(ctx, vol.Name, true)
		s.logger.ErrorContext(ctx, "Failed to pull Velocity image", "image", VelocityImage, "error", err)
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}

	// Ensure minecraft network exists
	s.logger.DebugContext(ctx, "Ensuring minecraft network exists", "network", MinecraftNetworkName)
	if err := s.ensureNetwork(ctx, MinecraftNetworkName); err != nil {
		s.dockerService.client.VolumeRemove(ctx, vol.Name, true)
		s.logger.ErrorContext(ctx, "Failed to ensure network", "network", MinecraftNetworkName, "error", err)
		return nil, fmt.Errorf("failed to ensure network: %w", err)
	}

	// Create container configuration
	containerConfig := &container.Config{
		Image: VelocityImage,
		Env: []string{
			"TYPE=VELOCITY",
			"MEMORY=512M",
		},
		Labels: map[string]string{
			"minecraft-proxy": "main",
		},
		ExposedPorts: nat.PortSet{
			"25577/tcp": struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/server", vol.Name),
		},
		PortBindings: nat.PortMap{
			"25577/tcp": []nat.PortBinding{
				{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", DefaultProxyPort)},
			},
		},
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			MinecraftNetworkName: {
				Aliases: []string{"velocity-proxy", "proxy"},
			},
		},
	}

	// Create container
	s.logger.InfoContext(ctx, "Creating proxy container", "container_name", "mc-proxy-main")
	resp, err := s.dockerService.client.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		networkConfig,
		nil,
		"mc-proxy-main",
	)
	if err != nil {
		s.dockerService.client.VolumeRemove(ctx, vol.Name, true)
		s.logger.ErrorContext(ctx, "Failed to create proxy container", "error", err)
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	s.logger.DebugContext(ctx, "Container created successfully", "container_id", resp.ID)

	// Create proxy model
	proxy := &models.ProxyServer{
		ID:          models.SingleProxyID,
		Name:        "Main Proxy",
		ContainerID: resp.ID,
		VolumeID:    vol.Name,
		Status:      models.ProxyStatusCreating,
		Port:        DefaultProxyPort,
	}

	// Save to database
	s.logger.DebugContext(ctx, "Saving proxy to database", "proxy_id", proxy.ID)
	if err := s.proxyRepo.Create(proxy); err != nil {
		s.dockerService.client.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
		s.dockerService.client.VolumeRemove(ctx, vol.Name, true)
		s.logger.ErrorContext(ctx, "Failed to save proxy to database", "proxy_id", proxy.ID, "error", err)
		return nil, fmt.Errorf("failed to save proxy to database: %w", err)
	}

	// Start the proxy
	s.logger.InfoContext(ctx, "Starting proxy server", "proxy_id", proxy.ID)
	if err := s.StartProxy(ctx); err != nil {
		s.logger.ErrorContext(ctx, "Failed to start proxy", "proxy_id", proxy.ID, "error", err)
		return nil, fmt.Errorf("failed to start proxy: %w", err)
	}

	s.logger.InfoContext(ctx, "Proxy server created successfully", "proxy_id", proxy.ID, "container_id", resp.ID)
	return proxy, nil
}

// ensureNetwork creates the minecraft network if it doesn't exist
func (s *ProxyService) ensureNetwork(ctx context.Context, networkName string) error {
	networks, err := s.dockerService.client.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return err
	}

	for _, net := range networks {
		if net.Name == networkName {
			return nil
		}
	}

	// Create network
	_, err = s.dockerService.client.NetworkCreate(ctx, networkName, network.CreateOptions{
		Driver: "bridge",
		Labels: map[string]string{
			"minecraft-network": "true",
		},
	})
	return err
}

// StartProxy starts the proxy server
func (s *ProxyService) StartProxy(ctx context.Context) error {
	proxy, err := s.proxyRepo.FindByID(models.SingleProxyID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to find proxy", "error", err)
		return err
	}

	s.logger.InfoContext(ctx, "Starting proxy container", "proxy_id", proxy.ID, "container_id", proxy.ContainerID)

	if err := s.dockerService.client.ContainerStart(ctx, proxy.ContainerID, container.StartOptions{}); err != nil {
		s.logger.ErrorContext(ctx, "Failed to start proxy container", "proxy_id", proxy.ID, "container_id", proxy.ContainerID, "error", err)
		return fmt.Errorf("failed to start container: %w", err)
	}

	proxy.Status = models.ProxyStatusRunning
	if err := s.proxyRepo.Update(proxy); err != nil {
		s.logger.ErrorContext(ctx, "Failed to update proxy status", "proxy_id", proxy.ID, "error", err)
		return err
	}

	s.logger.InfoContext(ctx, "Proxy started successfully", "proxy_id", proxy.ID)
	return nil
}

// StopProxy stops the proxy server
func (s *ProxyService) StopProxy(ctx context.Context) error {
	proxy, err := s.proxyRepo.FindByID(models.SingleProxyID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to find proxy", "error", err)
		return err
	}

	s.logger.InfoContext(ctx, "Stopping proxy container", "proxy_id", proxy.ID, "container_id", proxy.ContainerID)

	timeout := 30
	if err := s.dockerService.client.ContainerStop(ctx, proxy.ContainerID, container.StopOptions{
		Timeout: &timeout,
	}); err != nil {
		s.logger.ErrorContext(ctx, "Failed to stop proxy container", "proxy_id", proxy.ID, "container_id", proxy.ContainerID, "error", err)
		return fmt.Errorf("failed to stop container: %w", err)
	}

	proxy.Status = models.ProxyStatusStopped
	if err := s.proxyRepo.Update(proxy); err != nil {
		s.logger.ErrorContext(ctx, "Failed to update proxy status", "proxy_id", proxy.ID, "error", err)
		return err
	}

	s.logger.InfoContext(ctx, "Proxy stopped successfully", "proxy_id", proxy.ID)
	return nil
}

// syncProxyState checks Docker container state and updates database if needed
func (s *ProxyService) syncProxyState(ctx context.Context, proxy *models.ProxyServer) error {
	// Get container state from Docker
	state, err := s.dockerService.GetContainerState(ctx, proxy.ContainerID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get container state during sync",
			"proxy_id", proxy.ID,
			"error", err)
		return fmt.Errorf("failed to get container state: %w", err)
	}

	// Determine the new status based on container state
	var newStatus models.ProxyStatus
	if !state.Exists {
		// Container doesn't exist anymore (deleted manually or crashed)
		s.logger.WarnContext(ctx, "Proxy container no longer exists in Docker, marking as stopped",
			"proxy_id", proxy.ID,
			"previous_status", proxy.Status,
			"container_id", proxy.ContainerID)
		newStatus = models.ProxyStatusStopped
		proxy.ContainerID = "" // Clear the container ID
	} else if state.Running {
		newStatus = models.ProxyStatusRunning
	} else if state.Restarting {
		newStatus = models.ProxyStatusCreating
	} else if state.Dead || state.OOMKilled {
		newStatus = models.ProxyStatusError
	} else {
		// Stopped, paused, or exited
		newStatus = models.ProxyStatusStopped
	}

	// Update database if status changed
	if newStatus != proxy.Status {
		s.logger.InfoContext(ctx, "Proxy status changed, updating database",
			"proxy_id", proxy.ID,
			"previous_status", proxy.Status,
			"new_status", newStatus)
		proxy.Status = newStatus
		if err := s.proxyRepo.Update(proxy); err != nil {
			s.logger.ErrorContext(ctx, "Failed to update proxy status in database",
				"proxy_id", proxy.ID,
				"error", err)
			return err
		}
	}

	return nil
}

// GetProxy retrieves the proxy with synced state
func (s *ProxyService) GetProxy(ctx context.Context) (*models.ProxyServer, error) {
	proxy, err := s.proxyRepo.FindByID(models.SingleProxyID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to retrieve proxy from database", "error", err)
		return nil, err
	}

	// Sync state before returning
	if err := s.syncProxyState(ctx, proxy); err != nil {
		// Return proxy with last known state if sync fails
		s.logger.WarnContext(ctx, "Failed to sync proxy state, returning last known state",
			"proxy_id", proxy.ID,
			"error", err)
	}

	return proxy, nil
}

// ConnectServerToProxy connects a server to the minecraft network so proxy can reach it
func (s *ProxyService) ConnectServerToProxy(ctx context.Context, server *models.MinecraftServer) error {
	s.logger.InfoContext(ctx, "Connecting server to proxy network", "server_id", server.ID, "server_name", server.Name)

	// Ensure network exists
	if err := s.ensureNetwork(ctx, MinecraftNetworkName); err != nil {
		s.logger.ErrorContext(ctx, "Failed to ensure network exists", "server_id", server.ID, "error", err)
		return err
	}

	// Check if already connected
	containerInfo, err := s.dockerService.client.ContainerInspect(ctx, server.ContainerID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to inspect server container", "server_id", server.ID, "container_id", server.ContainerID, "error", err)
		return err
	}

	for netName := range containerInfo.NetworkSettings.Networks {
		if netName == MinecraftNetworkName {
			s.logger.DebugContext(ctx, "Server already connected to network", "server_id", server.ID, "network", MinecraftNetworkName)
			return nil // Already connected
		}
	}

	// Connect to network with server name as alias
	s.logger.InfoContext(ctx, "Connecting server to network", "server_id", server.ID, "server_name", server.Name, "network", MinecraftNetworkName)
	if err := s.dockerService.client.NetworkConnect(ctx, MinecraftNetworkName, server.ContainerID, &network.EndpointSettings{
		Aliases: []string{server.Name},
	}); err != nil {
		s.logger.ErrorContext(ctx, "Failed to connect server to network", "server_id", server.ID, "network", MinecraftNetworkName, "error", err)
		return err
	}

	s.logger.InfoContext(ctx, "Server connected to network successfully", "server_id", server.ID, "network", MinecraftNetworkName)
	return nil
}

// RegenerateProxyConfig regenerates the Velocity configuration based on all servers
func (s *ProxyService) RegenerateProxyConfig(ctx context.Context) error {
	proxy, err := s.proxyRepo.FindByID(models.SingleProxyID)
	if err != nil {
		return err
	}

	servers, err := s.serverRepo.FindAll()
	if err != nil {
		return err
	}

	var defaultServerName string
	if proxy.DefaultServerID != "" {
		server, err := s.serverRepo.FindByID(proxy.DefaultServerID)
		if err != nil {
			return err
		}
		defaultServerName = server.Name
	} else {
		defaultServerName = ""
	}

	config := s.generateVelocityConfig(servers, defaultServerName)

	// Write config to the container
	if err := s.writeConfigToContainer(ctx, proxy.ContainerID, config); err != nil {
		return fmt.Errorf("failed to write config to container: %w", err)
	}

	return nil
}

// writeConfigToContainer writes the Velocity config to the container via docker exec
func (s *ProxyService) writeConfigToContainer(ctx context.Context, containerID, config string) error {
	// Create exec to write the config file
	// We use sh -c with cat to write the file
	execConfig := container.ExecOptions{
		Cmd:          []string{"sh", "-c", fmt.Sprintf("cat > /server/velocity.toml << 'VELOCITYEOF'\n%s\nVELOCITYEOF", config)},
		AttachStdout: true,
		AttachStderr: true,
	}

	execResp, err := s.dockerService.client.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create exec: %w", err)
	}

	// Start the exec
	attachResp, err := s.dockerService.client.ContainerExecAttach(ctx, execResp.ID, container.ExecStartOptions{})
	if err != nil {
		return fmt.Errorf("failed to attach to exec: %w", err)
	}
	defer attachResp.Close()

	// Wait for exec to complete and read any output
	output, err := io.ReadAll(attachResp.Reader)
	if err != nil {
		return fmt.Errorf("failed to read exec output: %w", err)
	}

	// Check if exec was successful
	inspectResp, err := s.dockerService.client.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		return fmt.Errorf("failed to inspect exec: %w", err)
	}

	if inspectResp.ExitCode != 0 {
		return fmt.Errorf("exec failed with exit code %d: %s", inspectResp.ExitCode, string(output))
	}

	return nil
}

// generateVelocityConfig generates Velocity TOML configuration
func (s *ProxyService) generateVelocityConfig(servers []*models.MinecraftServer, defaultServer string) string {
	var serverEntries []string
	var tryList []string

	for _, server := range servers {
		// Use server name as DNS name (Docker network alias)
		serverEntries = append(serverEntries, fmt.Sprintf(`
%s = "%s:25565"`, server.Name, server.Name))
		tryList = append(tryList, fmt.Sprintf(`"%s"`, server.Name))
	}

	var tryConfigProperty string
	if defaultServer != "" {
		tryConfigProperty = fmt.Sprintf(`"%s"`, defaultServer)
	} else {
		tryConfigProperty = strings.Join(tryList, ", ")
	}
	config := fmt.Sprintf(`# Velocity Configuration
# Auto-generated by dockermc-cloud-manager

config-version = "2.7"

bind = "0.0.0.0:25577"
motd = "<aqua>Minecraft Server Network</aqua>"
show-max-players = 500
online-mode = true
force-key-authentication = false

# Player information forwarding settings
player-info-forwarding-mode = "legacy"

[servers]%s

try = [%s]

[forced-hosts]

[advanced]
compression-threshold = 256
compression-level = -1
login-ratelimit = 3000
connection-timeout = 5000
read-timeout = 30000

[query]
enabled = false
`, strings.Join(serverEntries, ""), tryConfigProperty)

	return config
}
