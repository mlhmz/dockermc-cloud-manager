package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mlhmz/dockermc-cloud-manager/internal/models"
	"github.com/mlhmz/dockermc-cloud-manager/internal/service"
)

// ServerHandler handles HTTP requests for server management
type ServerHandler struct {
	mcService *service.MinecraftServerService
}

// NewServerHandler creates a new ServerHandler
func NewServerHandler(mcService *service.MinecraftServerService) *ServerHandler {
	return &ServerHandler{
		mcService: mcService,
	}
}

// CreateServer handles POST /api/v1/servers
func (h *ServerHandler) CreateServer(w http.ResponseWriter, r *http.Request) {
	var req models.CreateServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	server, err := h.mcService.CreateServer(r.Context(), &req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

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

	if err := h.mcService.DeleteServer(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// StartServer handles POST /api/v1/servers/{id}/start
func (h *ServerHandler) StartServer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Server ID is required")
		return
	}

	if err := h.mcService.StartServer(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

// StopServer handles POST /api/v1/servers/{id}/stop
func (h *ServerHandler) StopServer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Server ID is required")
		return
	}

	if err := h.mcService.StopServer(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

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
