package handlers

import (
	"bufio"
	"io"
	"log/slog"
	"net/http"

	"github.com/coder/websocket"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/mlhmz/dockermc-cloud-manager/internal/service"
)

// LogsHandler handles WebSocket connections for streaming server logs
type LogsHandler struct {
	mcService *service.MinecraftServerService
	logger    *slog.Logger
}

// NewLogsHandler creates a new LogsHandler
func NewLogsHandler(mcService *service.MinecraftServerService, logger *slog.Logger) *LogsHandler {
	return &LogsHandler{
		mcService: mcService,
		logger:    logger,
	}
}

// StreamLogs handles WebSocket connections for streaming server logs
func (h *LogsHandler) StreamLogs(w http.ResponseWriter, r *http.Request) {
	// Get server ID from path
	serverID := r.PathValue("id")
	if serverID == "" {
		http.Error(w, "Server ID is required", http.StatusBadRequest)
		return
	}

	h.logger.InfoContext(r.Context(), "WebSocket connection requested for server logs", "server_id", serverID)

	// Verify server exists
	server, err := h.mcService.GetServer(r.Context(), serverID)
	if err != nil {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	// Upgrade HTTP connection to WebSocket
	// nhooyr handles origin checking, compression, and accepts options
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// In production, configure InsecureSkipVerify: false and proper OriginPatterns
		InsecureSkipVerify: true,
	})
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to upgrade to WebSocket", "error", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	h.logger.InfoContext(r.Context(), "WebSocket connection established", "server_id", serverID)

	// Get query parameters for log options
	follow := r.URL.Query().Get("follow") != "false" // Default to true
	tail := r.URL.Query().Get("tail")
	if tail == "" {
		tail = "100" // Default to last 100 lines
	}

	// Create context with cancellation
	ctx := r.Context()

	// Start streaming logs
	logReader, err := h.mcService.GetServerLogs(ctx, server.ContainerID, follow, tail)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to get server logs", "server_id", serverID, "error", err)
		conn.Write(ctx, websocket.MessageText, []byte("Error: Failed to retrieve logs"))
		return
	}
	defer logReader.Close()

	// Stream logs using stdcopy to demultiplex Docker's format
	// Create a pipe to convert io.Writer to line-based WebSocket messages
	pr, pw := io.Pipe()
	defer pr.Close()

	// Start demultiplexing in a goroutine
	go func() {
		defer pw.Close()
		_, err := stdcopy.StdCopy(pw, pw, logReader)
		if err != nil && err != io.EOF {
			h.logger.ErrorContext(ctx, "Error demultiplexing logs", "server_id", serverID, "error", err)
		}
	}()

	// Read from pipe and send to WebSocket
	scanner := bufio.NewScanner(pr)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// Send log line to WebSocket client
		// nhooyr automatically handles ping/pong and connection keep-alive
		err := conn.Write(ctx, websocket.MessageText, line)
		if err != nil {
			h.logger.InfoContext(ctx, "Client disconnected or write error", "server_id", serverID, "error", err)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		h.logger.ErrorContext(ctx, "Error reading logs", "server_id", serverID, "error", err)
	}

	h.logger.InfoContext(ctx, "Log streaming completed", "server_id", serverID)
}
