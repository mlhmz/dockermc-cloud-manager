package handlers

import (
	"net/http"
	"os"
	"path/filepath"
)

// ServeOpenAPISpec serves the OpenAPI specification YAML file
func ServeOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	// Get the project root directory
	wd, err := os.Getwd()
	if err != nil {
		http.Error(w, "Failed to get working directory", http.StatusInternalServerError)
		return
	}

	// Construct path to OpenAPI spec
	specPath := filepath.Join(wd, "api", "openapi.yaml")

	// Read the file
	data, err := os.ReadFile(specPath)
	if err != nil {
		http.Error(w, "Failed to read OpenAPI specification", http.StatusInternalServerError)
		return
	}

	// Set appropriate headers
	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
