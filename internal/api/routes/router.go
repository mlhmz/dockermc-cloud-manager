package routes

import (
	"bufio"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/mlhmz/dockermc-cloud-manager/internal/api/handlers"
	"github.com/mlhmz/dockermc-cloud-manager/internal/service"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// NewRouter creates and configures the HTTP router
func NewRouter(mcService *service.MinecraftServerService, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", healthCheckHandler)

	// Initialize handlers
	serverHandler := handlers.NewServerHandler(mcService, logger)
	logsHandler := handlers.NewLogsHandler(mcService, logger)

	// Server management endpoints
	mux.HandleFunc("POST /api/v1/servers", serverHandler.CreateServer)
	mux.HandleFunc("GET /api/v1/servers", serverHandler.ListServers)
	mux.HandleFunc("GET /api/v1/servers/{id}", serverHandler.GetServer)
	mux.HandleFunc("DELETE /api/v1/servers/{id}", serverHandler.DeleteServer)
	mux.HandleFunc("POST /api/v1/servers/{id}/start", serverHandler.StartServer)
	mux.HandleFunc("POST /api/v1/servers/{id}/stop", serverHandler.StopServer)

	// WebSocket endpoints
	mux.HandleFunc("GET /api/v1/servers/{id}/logs", logsHandler.StreamLogs)

	// API Documentation endpoints
	mux.HandleFunc("GET /api/openapi.yaml", handlers.ServeOpenAPISpec)
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/api/openapi.yaml"),
	))

	// Apply middleware
	return loggingMiddleware(logger, corsMiddleware(mux))
}

// healthCheckHandler returns the health status of the API
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

// loggingMiddleware logs HTTP requests with structured logging
func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		logger.InfoContext(r.Context(),
			"HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration_ms", duration.Milliseconds(),
			"remote_addr", r.RemoteAddr,
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Hijack implements http.Hijacker interface for WebSocket support
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return h.Hijack()
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
