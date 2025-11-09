package handlers

import (
	"bufio"
	"context"
	"encoding/json"
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

// CommandMessage represents a command sent from the client
type CommandMessage struct {
	Type    string `json:"type"`    // "command"
	Command string `json:"command"` // The Minecraft command to execute
}

// ResponseMessage represents a response sent to the client
type ResponseMessage struct {
	Type    string `json:"type"`    // "log", "command_result", "error"
	Content string `json:"content"` // The message content
}

// StreamLogs handles WebSocket connections for streaming server logs and executing commands
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
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Start streaming logs
	logReader, err := h.mcService.GetServerLogs(ctx, server.ContainerID, follow, tail)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to get server logs", "server_id", serverID, "error", err)
		h.sendError(ctx, conn, "Failed to retrieve logs")
		return
	}
	defer logReader.Close()

	// Channel to signal when log streaming is done
	logsDone := make(chan struct{})

	// Start goroutine to read commands from client
	go h.handleClientMessages(ctx, conn, server.ContainerID, serverID, cancel)

	// Stream logs in a goroutine
	go func() {
		defer close(logsDone)
		h.streamLogs(ctx, conn, logReader, serverID)
	}()

	// Wait for log streaming to complete
	<-logsDone

	h.logger.InfoContext(ctx, "Log streaming completed", "server_id", serverID)
}

// handleClientMessages reads incoming WebSocket messages and handles commands
func (h *LogsHandler) handleClientMessages(ctx context.Context, conn *websocket.Conn, containerID, serverID string, cancel context.CancelFunc) {
	defer cancel() // Cancel context when client disconnects

	for {
		// Read message from client
		msgType, data, err := conn.Read(ctx)
		if err != nil {
			if ctx.Err() != nil {
				// Context was cancelled, normal shutdown
				return
			}
			h.logger.InfoContext(ctx, "Client disconnected", "server_id", serverID, "error", err)
			return
		}

		// Only handle text messages
		if msgType != websocket.MessageText {
			h.logger.WarnContext(ctx, "Received non-text message", "server_id", serverID, "type", msgType)
			continue
		}

		// Parse command message
		var cmdMsg CommandMessage
		if err := json.Unmarshal(data, &cmdMsg); err != nil {
			h.logger.ErrorContext(ctx, "Failed to parse command message", "server_id", serverID, "error", err)
			h.sendError(ctx, conn, "Invalid message format")
			continue
		}

		// Handle command execution
		if cmdMsg.Type == "command" {
			h.logger.InfoContext(ctx, "Executing command", "server_id", serverID, "command", cmdMsg.Command)

			output, err := h.mcService.ExecuteCommand(ctx, containerID, cmdMsg.Command)
			if err != nil {
				h.logger.ErrorContext(ctx, "Failed to execute command", "server_id", serverID, "command", cmdMsg.Command, "error", err)
				h.sendError(ctx, conn, "Failed to execute command: "+err.Error())
				continue
			}

			// Send command result back to client
			h.sendCommandResult(ctx, conn, output)
		}
	}
}

// streamLogs reads from the log reader and sends logs to the WebSocket client
func (h *LogsHandler) streamLogs(ctx context.Context, conn *websocket.Conn, logReader io.ReadCloser, serverID string) {
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
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		// Send log line to WebSocket client
		if err := h.sendLog(ctx, conn, line); err != nil {
			if ctx.Err() != nil {
				// Context was cancelled, normal shutdown
				return
			}
			h.logger.InfoContext(ctx, "Client disconnected or write error", "server_id", serverID, "error", err)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		h.logger.ErrorContext(ctx, "Error reading logs", "server_id", serverID, "error", err)
	}
}

// sendLog sends a log message to the WebSocket client
func (h *LogsHandler) sendLog(ctx context.Context, conn *websocket.Conn, content string) error {
	msg := ResponseMessage{
		Type:    "log",
		Content: content,
	}
	data, _ := json.Marshal(msg)
	return conn.Write(ctx, websocket.MessageText, data)
}

// sendCommandResult sends a command execution result to the WebSocket client
func (h *LogsHandler) sendCommandResult(ctx context.Context, conn *websocket.Conn, content string) error {
	msg := ResponseMessage{
		Type:    "command_result",
		Content: content,
	}
	data, _ := json.Marshal(msg)
	return conn.Write(ctx, websocket.MessageText, data)
}

// sendError sends an error message to the WebSocket client
func (h *LogsHandler) sendError(ctx context.Context, conn *websocket.Conn, content string) error {
	msg := ResponseMessage{
		Type:    "error",
		Content: content,
	}
	data, _ := json.Marshal(msg)
	return conn.Write(ctx, websocket.MessageText, data)
}
