package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/mlhmz/dockermc-cloud-manager/internal/models"
	"github.com/mlhmz/dockermc-cloud-manager/internal/service"
)

// ServerHandler handles HTTP requests for server management
type ServerHandler struct {
	mcService *service.MinecraftServerService
	logger    *slog.Logger
}

// NewServerHandler creates a new ServerHandler
func NewServerHandler(mcService *service.MinecraftServerService, logger *slog.Logger) *ServerHandler {
	return &ServerHandler{
		mcService: mcService,
		logger:    logger,
	}
}

// CreateServer handles POST /api/v1/servers
func (h *ServerHandler) CreateServer(w http.ResponseWriter, r *http.Request) {
	var req models.CreateServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WarnContext(r.Context(), "Invalid request body for server creation", "error", err)
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	h.logger.InfoContext(r.Context(), "Creating new server", "name", req.Name)

	server, err := h.mcService.CreateServer(r.Context(), &req)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to create server", "name", req.Name, "error", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.logger.InfoContext(r.Context(), "Server created successfully", "id", server.ID, "name", server.Name)
	respondJSON(w, http.StatusCreated, server)
}

// ListServers handles GET /api/v1/servers
func (h *ServerHandler) ListServers(w http.ResponseWriter, r *http.Request) {
	servers, err := h.mcService.ListServers(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, servers)
}

// GetServer handles GET /api/v1/servers/{id}
func (h *ServerHandler) GetServer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Server ID is required")
		return
	}

	server, err := h.mcService.GetServer(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Server not found")
		return
	}

	respondJSON(w, http.StatusOK, server)
}

// DeleteServer handles DELETE /api/v1/servers/{id}
func (h *ServerHandler) DeleteServer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Server ID is required")
		return
	}

	h.logger.InfoContext(r.Context(), "Deleting server", "id", id)

	if err := h.mcService.DeleteServer(r.Context(), id); err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to delete server", "id", id, "error", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.logger.InfoContext(r.Context(), "Server deleted successfully", "id", id)
	w.WriteHeader(http.StatusNoContent)
}

// StartServer handles POST /api/v1/servers/{id}/start
func (h *ServerHandler) StartServer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Server ID is required")
		return
	}

	h.logger.InfoContext(r.Context(), "Starting server", "id", id)

	if err := h.mcService.StartServer(r.Context(), id); err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to start server", "id", id, "error", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.logger.InfoContext(r.Context(), "Server started successfully", "id", id)
	respondJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

// StopServer handles POST /api/v1/servers/{id}/stop
func (h *ServerHandler) StopServer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Server ID is required")
		return
	}

	h.logger.InfoContext(r.Context(), "Stopping server", "id", id)

	if err := h.mcService.StopServer(r.Context(), id); err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to stop server", "id", id, "error", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.logger.InfoContext(r.Context(), "Server stopped successfully", "id", id)
	respondJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

// Helper functions

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
