package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Port           int
	DockerNetwork  string
	VelocityImage  string
	MinecraftImage string
	DatabasePath   string
}

// Load reads configuration from environment variables with defaults
func Load() (*Config, error) {
	port := 8080
	if envPort := os.Getenv("API_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	dockerNetwork := os.Getenv("DOCKER_NETWORK")
	if dockerNetwork == "" {
		dockerNetwork = "minecraft-network"
	}

	velocityImage := os.Getenv("VELOCITY_IMAGE")
	if velocityImage == "" {
		velocityImage = "itzg/bungeecord:latest"
	}

	minecraftImage := os.Getenv("MINECRAFT_IMAGE")
	if minecraftImage == "" {
		minecraftImage = "itzg/minecraft-server:latest"
	}

	databasePath := os.Getenv("DATABASE_PATH")
	if databasePath == "" {
		databasePath = "./data/dockermc.db"
	}

	return &Config{
		Port:           port,
		DockerNetwork:  dockerNetwork,
		VelocityImage:  velocityImage,
		MinecraftImage: minecraftImage,
		DatabasePath:   databasePath,
	}, nil
}
