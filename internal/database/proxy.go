package database

import (
	"fmt"
	"log/slog"

	"github.com/mlhmz/dockermc-cloud-manager/internal/models"
	"gorm.io/gorm"
)

// ProxyRepository provides database operations for ProxyServer
type ProxyRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewProxyRepository creates a new proxy repository
func NewProxyRepository(db *DB) *ProxyRepository {
	return &ProxyRepository{
		db:     db.DB,
		logger: db.logger,
	}
}

// Create inserts a new proxy into the database
func (r *ProxyRepository) Create(proxy *models.ProxyServer) error {
	result := r.db.Create(proxy)
	if result.Error != nil {
		r.logger.Error("Failed to create proxy in database", "error", result.Error)
		return result.Error
	}
	r.logger.Debug("Proxy created in database", "id", proxy.ID, "name", proxy.Name)
	return nil
}

// FindByID retrieves a proxy by its ID
func (r *ProxyRepository) FindByID(id string) (*models.ProxyServer, error) {
	var proxy models.ProxyServer
	result := r.db.First(&proxy, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("proxy not found")
		}
		r.logger.Error("Failed to find proxy by ID", "id", id, "error", result.Error)
		return nil, result.Error
	}
	return &proxy, nil
}

// FindByName retrieves a proxy by its name
func (r *ProxyRepository) FindByName(name string) (*models.ProxyServer, error) {
	var proxy models.ProxyServer
	result := r.db.First(&proxy, "name = ?", name)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("proxy not found")
		}
		r.logger.Error("Failed to find proxy by name", "name", name, "error", result.Error)
		return nil, result.Error
	}
	return &proxy, nil
}

// FindAll retrieves all proxies
func (r *ProxyRepository) FindAll() ([]*models.ProxyServer, error) {
	var proxies []*models.ProxyServer
	result := r.db.Find(&proxies)
	if result.Error != nil {
		r.logger.Error("Failed to find all proxies", "error", result.Error)
		return nil, result.Error
	}
	return proxies, nil
}

// Update updates a proxy in the database
func (r *ProxyRepository) Update(proxy *models.ProxyServer) error {
	result := r.db.Save(proxy)
	if result.Error != nil {
		r.logger.Error("Failed to update proxy", "id", proxy.ID, "error", result.Error)
		return result.Error
	}
	r.logger.Debug("Proxy updated in database", "id", proxy.ID, "name", proxy.Name)
	return nil
}

// Delete removes a proxy from the database
func (r *ProxyRepository) Delete(id string) error {
	result := r.db.Unscoped().Delete(&models.ProxyServer{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("Failed to delete proxy", "id", id, "error", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("proxy not found")
	}
	r.logger.Debug("Proxy deleted from database", "id", id)
	return nil
}
