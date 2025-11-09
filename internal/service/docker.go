package service

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

// DockerService handles Docker operations
type DockerService struct {
	client *client.Client
}

// NewDockerService creates a new Docker service
func NewDockerService() (*DockerService, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &DockerService{
		client: cli,
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
		log.Printf("Image %s already exists locally", imageName)
		return nil
	}

	// Image doesn't exist, pull it
	log.Printf("Pulling image %s...", imageName)
	reader, err := s.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	defer reader.Close()

	// Read the pull output to ensure it completes
	// This is necessary because ImagePull is asynchronous
	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		return fmt.Errorf("error reading image pull output: %w", err)
	}

	log.Printf("Successfully pulled image %s", imageName)
	return nil
}
