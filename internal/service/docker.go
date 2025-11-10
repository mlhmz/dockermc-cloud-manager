package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

// DockerService handles Docker operations
type DockerService struct {
	client *client.Client
	logger *slog.Logger
}

// NewDockerService creates a new Docker service
func NewDockerService(logger *slog.Logger) (*DockerService, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &DockerService{
		client: cli,
		logger: logger,
	}, nil
}

// Close closes the Docker client connection
func (s *DockerService) Close() error {
	return s.client.Close()
}

// Ping checks if Docker daemon is accessible
func (s *DockerService) Ping(ctx context.Context) error {
	_, err := s.client.Ping(ctx)
	return err
}

// GetClient returns the underlying Docker client
func (s *DockerService) GetClient() *client.Client {
	return s.client
}

// PullImage pulls a Docker image if it doesn't exist locally
func (s *DockerService) PullImage(ctx context.Context, imageName string) error {
	// Check if image already exists
	_, _, err := s.client.ImageInspectWithRaw(ctx, imageName)
	if err == nil {
		// Image already exists
		s.logger.InfoContext(ctx, "Image already exists locally", "image", imageName)
		return nil
	}

	// Image doesn't exist, pull it
	s.logger.InfoContext(ctx, "Pulling Docker image", "image", imageName)
	reader, err := s.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to pull image", "image", imageName, "error", err)
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	defer reader.Close()

	// Read the pull output to ensure it completes
	// This is necessary because ImagePull is asynchronous
	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		s.logger.ErrorContext(ctx, "Error reading image pull output", "image", imageName, "error", err)
		return fmt.Errorf("error reading image pull output: %w", err)
	}

	s.logger.InfoContext(ctx, "Successfully pulled image", "image", imageName)
	return nil
}

// ContainerState represents the state of a Docker container
type ContainerState struct {
	Exists      bool // Whether the container exists in Docker
	Running     bool // Whether the container is running
	Restarting  bool // Whether the container is restarting
	Dead        bool // Whether the container is dead
	OOMKilled   bool // Whether the container was killed due to OOM
}

// GetContainerState inspects a container and returns its current state
func (s *DockerService) GetContainerState(ctx context.Context, containerID string) (*ContainerState, error) {
	if containerID == "" {
		return &ContainerState{Exists: false}, nil
	}

	// Inspect the container
	containerJSON, err := s.client.ContainerInspect(ctx, containerID)
	if err != nil {
		// Container doesn't exist
		return &ContainerState{Exists: false}, nil
	}

	// Container exists, extract state information
	return &ContainerState{
		Exists:     true,
		Running:    containerJSON.State.Running,
		Restarting: containerJSON.State.Restarting,
		Dead:       containerJSON.State.Dead,
		OOMKilled:  containerJSON.State.OOMKilled,
	}, nil
}
