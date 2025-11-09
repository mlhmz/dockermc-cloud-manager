package models

import (
	"time"

	"gorm.io/gorm"
)

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
	ID          string         `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"uniqueIndex;not null"`
	ContainerID string         `json:"container_id" gorm:"index"`
	VolumeID    string         `json:"volume_id"`
	Status      ServerStatus   `json:"status" gorm:"type:varchar(20)"`
	Port        int            `json:"port"`
	MaxPlayers  int            `json:"max_players" gorm:"not null"`
	MOTD        string         `json:"motd"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
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
