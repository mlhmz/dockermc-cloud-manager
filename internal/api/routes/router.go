package routes

import (
	"encoding/json"
	"net/http"

	"github.com/mlhmz/dockermc-cloud-manager/internal/api/handlers"
	"github.com/mlhmz/dockermc-cloud-manager/internal/service"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// NewRouter creates and configures the HTTP router
func NewRouter(mcService *service.MinecraftServerService) http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", healthCheckHandler)

	// Initialize handlers
	serverHandler := handlers.NewServerHandler(mcService)

	// Server management endpoints
	mux.HandleFunc("POST /api/v1/servers", serverHandler.CreateServer)
	mux.HandleFunc("GET /api/v1/servers", serverHandler.ListServers)
	mux.HandleFunc("GET /api/v1/servers/{id}", serverHandler.GetServer)
	mux.HandleFunc("DELETE /api/v1/servers/{id}", serverHandler.DeleteServer)
	mux.HandleFunc("POST /api/v1/servers/{id}/start", serverHandler.StartServer)
	mux.HandleFunc("POST /api/v1/servers/{id}/stop", serverHandler.StopServer)

	// API Documentation endpoints
	mux.HandleFunc("GET /api/openapi.yaml", handlers.ServeOpenAPISpec)
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/api/openapi.yaml"),
	))

	// Apply middleware
	return loggingMiddleware(corsMiddleware(mux))
}

// healthCheckHandler returns the health status of the API
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Replace with proper logger
		println(r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
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
