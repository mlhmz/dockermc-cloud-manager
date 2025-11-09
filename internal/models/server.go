package models

import "time"

// ServerStatus represents the current status of a Minecraft server
type ServerStatus string

const (
	StatusCreating ServerStatus = "creating"
	StatusRunning  ServerStatus = "running"
	StatusStopped  ServerStatus = "stopped"
	StatusError    ServerStatus = "error"
)

// MinecraftServer represents a Minecraft server instance
type MinecraftServer struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	ContainerID string       `json:"container_id"`
	VolumeID    string       `json:"volume_id"`
	Status      ServerStatus `json:"status"`
	Port        int          `json:"port"`
	MaxPlayers  int          `json:"max_players"`
	MOTD        string       `json:"motd"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// CreateServerRequest represents the request body for creating a new server
type CreateServerRequest struct {
	Name       string `json:"name" binding:"required"`
	MaxPlayers int    `json:"max_players"`
	MOTD       string `json:"motd"`
	Version    string `json:"version"`
}

// UpdateServerRequest represents the request body for updating a server
type UpdateServerRequest struct {
	MaxPlayers *int    `json:"max_players,omitempty"`
	MOTD       *string `json:"motd,omitempty"`
}
