package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/mlhmz/dockermc-cloud-manager/internal/service"
)

// ProxyHandler handles HTTP requests for proxy operations
type ProxyHandler struct {
	proxyService *service.ProxyService
	logger       *slog.Logger
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(proxyService *service.ProxyService, logger *slog.Logger) *ProxyHandler {
	return &ProxyHandler{
		proxyService: proxyService,
		logger:       logger,
	}
}

// GetProxy retrieves the proxy status
func (h *ProxyHandler) GetProxy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	proxy, err := h.proxyService.GetProxy(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to get proxy", "error", err)
		http.Error(w, "Proxy not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(proxy)
}

// StartProxy starts the proxy
func (h *ProxyHandler) StartProxy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Ensure proxy exists
	proxy, err := h.proxyService.EnsureProxyExists(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to ensure proxy exists", "error", err)
		http.Error(w, "Failed to create/find proxy", http.StatusInternalServerError)
		return
	}

	// Start the proxy
	if err := h.proxyService.StartProxy(ctx); err != nil {
		h.logger.ErrorContext(ctx, "Failed to start proxy", "error", err)
		http.Error(w, "Failed to start proxy", http.StatusInternalServerError)
		return
	}

	// Regenerate config to include all servers
	if err := h.proxyService.RegenerateProxyConfig(ctx); err != nil {
		h.logger.WarnContext(ctx, "Failed to regenerate proxy config", "error", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(proxy)
}

// StopProxy stops the proxy
func (h *ProxyHandler) StopProxy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.proxyService.StopProxy(ctx); err != nil {
		h.logger.ErrorContext(ctx, "Failed to stop proxy", "error", err)
		http.Error(w, "Failed to stop proxy", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Proxy stopped successfully",
	})
}

// RegenerateConfig regenerates the proxy configuration
func (h *ProxyHandler) RegenerateConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.proxyService.RegenerateProxyConfig(ctx); err != nil {
		h.logger.ErrorContext(ctx, "Failed to regenerate proxy config", "error", err)
		http.Error(w, "Failed to regenerate configuration", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Configuration regenerated successfully",
	})
}
