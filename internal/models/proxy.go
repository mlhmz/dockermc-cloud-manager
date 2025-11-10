package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	// SingleProxyID is the ID of the single Velocity proxy instance
	SingleProxyID = "main-proxy"
)

// ProxyServer represents the single Velocity proxy server instance
type ProxyServer struct {
	ID              string          `json:"id" gorm:"primaryKey"`
	Name            string          `json:"name" gorm:"not null"`
	ContainerID     string          `json:"container_id" gorm:"index"`
	VolumeID        string          `json:"volume_id"`
	DefaultServerID string          `json:"default_server_id"`
	Status          ContainerStatus `json:"status" gorm:"type:varchar(20)"`
	Port            int             `json:"port" gorm:"not null"` // Public port (typically 25565)
	CreatedAt       time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt  `json:"-" gorm:"index"`
}

type UpdateProxyRequest struct {
	DefaultServerID string `json:"default_server_id"`
}
