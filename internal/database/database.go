package database

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/mlhmz/dockermc-cloud-manager/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB wraps the GORM database connection
type DB struct {
	*gorm.DB
	logger *slog.Logger
}

// New creates a new database connection
func New(dbPath string, log *slog.Logger) (*DB, error) {
	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Configure GORM logger to be quiet (we use slog instead)
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	// Open database connection
	db, err := gorm.Open(sqlite.Open(dbPath), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Info("Database connection established", "path", dbPath)

	// Auto-migrate schemas
	if err := db.AutoMigrate(&models.MinecraftServer{}); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate schemas: %w", err)
	}

	log.Info("Database schemas migrated successfully")

	return &DB{
		DB:     db,
		logger: log,
	}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// ServerRepository provides database operations for MinecraftServer
type ServerRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewServerRepository creates a new server repository
func NewServerRepository(db *DB) *ServerRepository {
	return &ServerRepository{
		db:     db.DB,
		logger: db.logger,
	}
}

// Create inserts a new server into the database
func (r *ServerRepository) Create(server *models.MinecraftServer) error {
	result := r.db.Create(server)
	if result.Error != nil {
		r.logger.Error("Failed to create server in database", "error", result.Error)
		return result.Error
	}
	r.logger.Debug("Server created in database", "id", server.ID, "name", server.Name)
	return nil
}

// FindByID retrieves a server by its ID
func (r *ServerRepository) FindByID(id string) (*models.MinecraftServer, error) {
	var server models.MinecraftServer
	result := r.db.First(&server, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("server not found")
		}
		r.logger.Error("Failed to find server by ID", "id", id, "error", result.Error)
		return nil, result.Error
	}
	return &server, nil
}

// FindByName retrieves a server by its name
func (r *ServerRepository) FindByName(name string) (*models.MinecraftServer, error) {
	var server models.MinecraftServer
	result := r.db.First(&server, "name = ?", name)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("server not found")
		}
		r.logger.Error("Failed to find server by name", "name", name, "error", result.Error)
		return nil, result.Error
	}
	return &server, nil
}

// FindAll retrieves all servers
func (r *ServerRepository) FindAll() ([]*models.MinecraftServer, error) {
	var servers []*models.MinecraftServer
	result := r.db.Find(&servers)
	if result.Error != nil {
		r.logger.Error("Failed to find all servers", "error", result.Error)
		return nil, result.Error
	}
	return servers, nil
}

// Update updates a server in the database
func (r *ServerRepository) Update(server *models.MinecraftServer) error {
	result := r.db.Save(server)
	if result.Error != nil {
		r.logger.Error("Failed to update server", "id", server.ID, "error", result.Error)
		return result.Error
	}
	r.logger.Debug("Server updated in database", "id", server.ID, "name", server.Name)
	return nil
}

// Delete removes a server from the database (soft delete)
func (r *ServerRepository) Delete(id string) error {
	result := r.db.Delete(&models.MinecraftServer{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("Failed to delete server", "id", id, "error", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("server not found")
	}
	r.logger.Debug("Server deleted from database", "id", id)
	return nil
}

// HardDelete permanently removes a server from the database
func (r *ServerRepository) HardDelete(id string) error {
	result := r.db.Unscoped().Delete(&models.MinecraftServer{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("Failed to hard delete server", "id", id, "error", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("server not found")
	}
	r.logger.Debug("Server permanently deleted from database", "id", id)
	return nil
}
