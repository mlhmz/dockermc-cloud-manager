# DockerMC Cloud Manager - Backend Documentation

## Project Overview

DockerMC Cloud Manager is a cloud management platform for orchestrating multiple Minecraft servers using Docker containers. The backend is built in Go with a clean, layered architecture following domain-driven design principles.

**Tech Stack:**
- Go 1.23+
- GORM (ORM) with SQLite
- Docker SDK for Go
- Cobra (CLI framework)
- WebSocket (coder/websocket)
- Structured logging (slog)

**Key Features:**
- Multi-server Minecraft orchestration
- Velocity proxy management
- Real-time log streaming via WebSocket
- Auto-configuration and networking
- Graceful state synchronization

## Architecture Overview

### Directory Structure

```
.
├── cmd/
│   └── api/
│       └── main.go                    # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/                  # HTTP request handlers
│   │   │   ├── server.go             # Minecraft server endpoints
│   │   │   ├── proxy.go              # Velocity proxy endpoints
│   │   │   ├── logs.go               # WebSocket log streaming
│   │   │   └── swagger.go            # OpenAPI spec serving
│   │   └── routes/
│   │       └── router.go             # HTTP routing and middleware
│   ├── cli/
│   │   └── commands/                  # Cobra CLI commands
│   │       ├── root.go               # Root command setup
│   │       ├── serve.go              # API server command
│   │       └── server.go             # Server management CLI
│   ├── config/
│   │   ├── config.go                 # Configuration loading
│   │   └── logger.go                 # Structured logging setup
│   ├── database/
│   │   ├── database.go               # GORM connection & server repo
│   │   └── proxy.go                  # Proxy repository
│   ├── models/
│   │   ├── server.go                 # MinecraftServer model
│   │   ├── proxy.go                  # ProxyServer model
│   │   └── status.go                 # Container status enum
│   └── service/
│       ├── docker.go                  # Docker client wrapper
│       ├── minecraft.go               # Minecraft server business logic
│       └── proxy.go                   # Velocity proxy business logic
├── api/
│   ├── openapi.yaml                   # OpenAPI 3.0 specification
│   └── .bruno/                        # Bruno API test collections
├── data/                              # SQLite database location
├── go.mod                             # Go dependencies
└── .env                               # Environment configuration
```

## Core Architecture Patterns

### 1. Layered Architecture

```
┌─────────────────────────────────────┐
│         HTTP Handlers               │  ← Request/Response, JSON marshaling
├─────────────────────────────────────┤
│         Service Layer               │  ← Business logic, orchestration
├─────────────────────────────────────┤
│    Repository Layer (GORM)          │  ← Database operations
├─────────────────────────────────────┤
│    Docker Service (SDK)             │  ← Container management
└─────────────────────────────────────┘
```

**Layers:**
1. **Handlers** - HTTP interface, validation, error responses
2. **Services** - Business logic, orchestration, state management
3. **Repositories** - Database CRUD operations
4. **Docker Service** - Container lifecycle management

### 2. Dependency Injection

Constructor-based dependency injection for loose coupling:

```go
func NewMinecraftServerService(
    dockerService *DockerService,
    repo *database.ServerRepository,
    logger *slog.Logger,
) *MinecraftServerService
```

**Circular Dependency Handling:**
```go
// In serve.go - bidirectional dependency
mcService := service.NewMinecraftServerService(dockerService, serverRepo, logger)
proxyService := service.NewProxyService(dockerService, proxyRepo, serverRepo, logger)
mcService.SetProxyService(proxyService)  // Inject after creation
```

### 3. Repository Pattern

Each entity gets its own repository with standard CRUD operations:

```go
type ServerRepository struct {
    db     *gorm.DB
    logger *slog.Logger
}

func (r *ServerRepository) Create(server *models.MinecraftServer) error
func (r *ServerRepository) FindByID(id string) (*models.MinecraftServer, error)
func (r *ServerRepository) FindByName(name string) (*models.MinecraftServer, error)
func (r *ServerRepository) FindAll() ([]models.MinecraftServer, error)
func (r *ServerRepository) Update(server *models.MinecraftServer) error
func (r *ServerRepository) Delete(id string) error
```

### 4. Singleton Pattern

Single proxy server instance (ID: `"main-proxy"`):

```go
const SingleProxyID = "main-proxy"

func (s *ProxyService) EnsureProxyExists(ctx context.Context) (*models.ProxyServer, error)
```

Proxy is created lazily on first access.

## Data Models

### MinecraftServer
```go
type MinecraftServer struct {
    ID          string          `json:"id" gorm:"primaryKey"`
    Name        string          `json:"name" gorm:"uniqueIndex;not null"`
    ContainerID string          `json:"container_id" gorm:"index"`
    VolumeID    string          `json:"volume_id"`
    Status      ContainerStatus `json:"status" gorm:"type:varchar(20)"`
    Port        int             `json:"port"`
    MaxPlayers  int             `json:"max_players" gorm:"not null"`
    MOTD        string          `json:"motd"`
    CreatedAt   time.Time       `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt   time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}
```

### ProxyServer
```go
type ProxyServer struct {
    ID              string          `json:"id" gorm:"primaryKey"`
    Name            string          `json:"name" gorm:"not null"`
    ContainerID     string          `json:"container_id" gorm:"index"`
    VolumeID        string          `json:"volume_id"`
    DefaultServerID string          `json:"default_server_id"`
    Status          ContainerStatus `json:"status" gorm:"type:varchar(20)"`
    Port            int             `json:"port" gorm:"not null"`
    CreatedAt       time.Time       `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt       time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}
```

### ContainerStatus
```go
type ContainerStatus string

const (
    StatusCreating ContainerStatus = "creating"
    StatusRunning  ContainerStatus = "running"
    StatusStopped  ContainerStatus = "stopped"
    StatusError    ContainerStatus = "error"
)
```

**Design Notes:**
- No foreign keys (loose coupling between servers and proxy)
- Unique constraint on server names (used as Docker network aliases)
- Indexed ContainerID for fast Docker state lookups
- Auto-timestamps via GORM

## Application Initialization

### Entry Point (cmd/api/main.go)

```go
func main() {
    // Load .env file
    err := godotenv.Load()

    // Execute CLI commands
    if err := commands.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### Serve Command Initialization

The `serve` command initializes the application in this order:

1. **Load Configuration** - Environment variables with defaults
2. **Setup Logger** - Structured logging (slog)
3. **Initialize Database** - SQLite with auto-migration
4. **Create Docker Service** - Connect to Docker daemon
5. **Create Repositories** - Server and Proxy repos
6. **Create Services** - Business logic layer with bidirectional deps
7. **Setup Router** - HTTP handlers with middleware
8. **Start Server** - HTTP server with graceful shutdown

```go
// Graceful shutdown
shutdown := make(chan os.Signal, 1)
signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

go func() {
    <-shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()
    srv.Shutdown(ctx)
}()
```

## HTTP API

### Routing

Uses Go 1.22+ native HTTP router with method-based routing:

```go
// Servers
mux.HandleFunc("POST /api/v1/servers", serverHandler.CreateServer)
mux.HandleFunc("GET /api/v1/servers", serverHandler.ListServers)
mux.HandleFunc("GET /api/v1/servers/{id}", serverHandler.GetServer)
mux.HandleFunc("DELETE /api/v1/servers/{id}", serverHandler.DeleteServer)
mux.HandleFunc("POST /api/v1/servers/{id}/start", serverHandler.StartServer)
mux.HandleFunc("POST /api/v1/servers/{id}/stop", serverHandler.StopServer)
mux.HandleFunc("GET /api/v1/servers/{id}/logs", logsHandler.StreamLogs)

// Proxy
mux.HandleFunc("GET /api/v1/proxy", proxyHandler.GetProxy)
mux.HandleFunc("PATCH /api/v1/proxy", proxyHandler.UpdateProxy)
mux.HandleFunc("POST /api/v1/proxy/start", proxyHandler.StartProxy)
mux.HandleFunc("POST /api/v1/proxy/stop", proxyHandler.StopProxy)
mux.HandleFunc("POST /api/v1/proxy/regenerate-config", proxyHandler.RegenerateConfig)

// Docs
mux.HandleFunc("GET /api/openapi.yaml", swaggerHandler.ServeOpenAPISpec)
mux.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)
```

### Middleware Chain

```go
loggingMiddleware(logger, corsMiddleware(mux))
```

**Order:** Logging → CORS → Handler

**CORS Middleware:**
```go
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

**Logging Middleware:**
- Logs method, path, status, duration, remote address
- Custom responseWriter to capture status codes
- Implements `http.Hijacker` for WebSocket support
- Context-aware logging with slog

### Handler Pattern

**Thin handlers** - delegate to service layer:

```go
func (h *ServerHandler) CreateServer(w http.ResponseWriter, r *http.Request) {
    var req CreateServerRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    server, err := h.service.CreateServer(r.Context(), &req)
    if err != nil {
        h.logger.ErrorContext(r.Context(), "Failed to create server", "error", err)
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondJSON(w, http.StatusCreated, server)
}
```

**Standard Helpers:**
```go
func respondJSON(w http.ResponseWriter, status int, data interface{})
func respondError(w http.ResponseWriter, status int, message string)
```

## Service Layer

### MinecraftServerService

**Key Responsibilities:**
- Server lifecycle management
- Docker container orchestration
- State synchronization
- Auto-configuration for proxy mode

**Server Creation Flow:**

1. Generate UUID for server ID
2. Create Docker volume (`mc-server-{id}`)
3. Pull Minecraft image (`itzg/minecraft-server:latest`)
4. Check if proxy exists
5. Configure container:
   - Environment: EULA=TRUE, MAX_PLAYERS, MOTD, VERSION, TYPE=PAPER
   - If proxy exists: ONLINE_MODE=FALSE, PATCH_DEFINITIONS
   - Restart policy: unless-stopped
   - Volume mount: /data
6. Create and start container
7. If proxy exists: Write BungeeCord patch to volume
8. Save server to database
9. Connect to proxy network and regenerate config
10. On any error: cleanup container and volume

**BungeeCord/Velocity Forwarding:**

Patch file written to volume before server starts:
```json
{
  "file": "/data/spigot.yml",
  "ops": [
    {
      "$set": {
        "path": "$.settings.bungeecord",
        "value": true,
        "value-type": "bool"
      }
    }
  ]
}
```

Uses temporary Alpine container to write file.

**State Synchronization:**

Before every read operation, sync database with Docker:
```go
func (s *MinecraftServerService) syncServerState(ctx context.Context, server *models.MinecraftServer) error {
    state := s.dockerService.GetContainerState(ctx, server.ContainerID)

    if !state.Exists {
        server.ContainerID = ""
        server.Status = models.StatusStopped
    } else if state.Running {
        server.Status = models.StatusRunning
    } else {
        server.Status = models.StatusStopped
    }

    return s.repo.Update(server)
}
```

**Why?** Database and Docker can diverge due to:
- Manual container deletion
- Docker daemon restarts
- Container crashes

### ProxyService

**Singleton Pattern:** One proxy instance (`main-proxy`)

**Lazy Initialization:**
```go
func (s *ProxyService) EnsureProxyExists(ctx context.Context) (*models.ProxyServer, error) {
    proxy, err := s.proxyRepo.FindByID(SingleProxyID)
    if err != nil {
        return s.createProxy(ctx)
    }
    return proxy, nil
}
```

**Proxy Creation Flow:**

1. Create Docker volume (`mc-proxy-main`)
2. Pull Velocity image (`itzg/bungeecord:latest`)
3. Ensure Docker network exists (`minecraft-network`)
4. Configure container:
   - Environment: TYPE=VELOCITY, MEMORY=512M
   - Port: 0.0.0.0:25565 → 25577/tcp
   - Network aliases: velocity-proxy, proxy
5. Save to database
6. Start container
7. On error: cleanup

**Network Management:**

```go
func (s *ProxyService) ensureNetwork(ctx context.Context) error {
    // Check if network exists
    networks, _ := s.dockerService.client.NetworkList(ctx, ...)
    for _, net := range networks {
        if net.Name == networkName {
            return nil
        }
    }

    // Create bridge network
    _, err := s.dockerService.client.NetworkCreate(ctx, networkName, ...)
    return err
}
```

**Configuration Generation:**

Generates Velocity TOML dynamically based on servers:
```go
func (s *ProxyService) RegenerateProxyConfig(ctx context.Context) error {
    // Get all servers
    servers, _ := s.serverRepo.FindAll()

    // Build config
    config := generateVelocityConfig(servers, proxy.DefaultServerID)

    // Write to container via exec
    cmd := []string{"sh", "-c", fmt.Sprintf("cat > /server/velocity.toml << 'EOF'\n%s\nEOF", config)}
    exec, _ := s.dockerService.client.ContainerExecCreate(ctx, proxy.ContainerID, ...)
    return s.dockerService.client.ContainerExecStart(ctx, exec.ID, ...)
}
```

**Connect Server to Network:**
```go
func (s *ProxyService) ConnectServerToProxy(ctx context.Context, server *models.MinecraftServer) error {
    // Check if already connected
    inspect, _ := s.dockerService.client.ContainerInspect(ctx, server.ContainerID)
    if _, exists := inspect.NetworkSettings.Networks[s.config.DockerNetwork]; exists {
        return nil  // Already connected
    }

    // Connect with server name as DNS alias
    return s.dockerService.client.NetworkConnect(ctx, s.config.DockerNetwork, server.ContainerID, &network.EndpointSettings{
        Aliases: []string{server.Name},
    })
}
```

### DockerService

**Core Wrapper** around Docker SDK:

```go
type DockerService struct {
    client *client.Client
    logger *slog.Logger
}

func NewDockerService(logger *slog.Logger) (*DockerService, error) {
    cli, err := client.NewClientWithOpts(
        client.FromEnv,
        client.WithAPIVersionNegotiation(),
    )
    // ...
}
```

**Key Methods:**
- `PullImage(ctx, image)` - Idempotent image pulling
- `GetContainerState(ctx, id)` - Get container state
- `Ping(ctx)` - Docker daemon health check

**Container State:**
```go
type ContainerState struct {
    Exists      bool
    Running     bool
    Restarting  bool
    Dead        bool
    OOMKilled   bool
}
```

## WebSocket Log Streaming

### Protocol

**Endpoint:** `GET /api/v1/servers/{id}/logs`

**Query Parameters:**
- `follow` (bool, default: true) - Follow logs in real-time
- `tail` (string, default: "100") - Number of lines to tail

**Message Types (JSON):**

Client → Server:
```json
{
  "type": "command",
  "command": "say Hello!"
}
```

Server → Client:
```json
{
  "type": "log",
  "content": "[12:34:56] [Server thread/INFO]: Starting server..."
}

{
  "type": "command_result",
  "content": "Command executed successfully"
}

{
  "type": "error",
  "content": "Failed to execute command"
}
```

### Implementation

**WebSocket Upgrade:**
```go
conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
    InsecureSkipVerify: true,  // Development
})
defer conn.Close(websocket.StatusNormalClosure, "")
```

**Concurrent Goroutines:**

1. **Read Loop** - Listen for client commands:
```go
func handleClientMessages(ctx context.Context, conn *websocket.Conn) {
    for {
        var msg CommandMessage
        err := wsjson.Read(ctx, conn, &msg)
        if err != nil {
            return
        }

        if msg.Type == "command" {
            result := executeCommand(ctx, containerID, msg.Command)
            wsjson.Write(ctx, conn, ResponseMessage{
                Type: "command_result",
                Content: result,
            })
        }
    }
}
```

2. **Write Loop** - Stream logs to client:
```go
func streamLogs(ctx context.Context, conn *websocket.Conn, logReader io.ReadCloser) {
    pr, pw := io.Pipe()
    go stdcopy.StdCopy(pw, pw, logReader)  // Demultiplex Docker logs

    scanner := bufio.NewScanner(pr)
    for scanner.Scan() {
        line := scanner.Text()
        err := wsjson.Write(ctx, conn, ResponseMessage{
            Type: "log",
            Content: line,
        })
        if err != nil {
            return
        }
    }
}
```

**Docker Log Demultiplexing:**

Docker multiplexes stdout/stderr. Use `stdcopy.StdCopy()` to demultiplex:
```go
import "github.com/docker/docker/pkg/stdcopy"

pr, pw := io.Pipe()
go stdcopy.StdCopy(pw, pw, dockerLogReader)
scanner := bufio.NewScanner(pr)
```

**Command Execution:**

Uses rcon-cli inside container:
```go
func executeCommand(ctx context.Context, containerID, command string) string {
    exec, _ := dockerClient.ContainerExecCreate(ctx, containerID, types.ExecConfig{
        Cmd: []string{"rcon-cli", command},
        AttachStdout: true,
        AttachStderr: true,
    })

    resp, _ := dockerClient.ContainerExecAttach(ctx, exec.ID, ...)
    defer resp.Close()

    output, _ := io.ReadAll(resp.Reader)
    return string(output)
}
```

## Configuration

### Environment Variables

```bash
# Server Configuration
API_PORT=8080                          # API server port

# Docker Configuration
DOCKER_NETWORK=minecraft-network       # Docker network name
VELOCITY_IMAGE=itzg/bungeecord:latest  # Velocity proxy image
MINECRAFT_IMAGE=itzg/minecraft-server:latest

# Database
DATABASE_PATH=./data/dockermc.db       # SQLite database path

# Logging
LOG_LEVEL=INFO                         # DEBUG, INFO, WARN, ERROR
LOG_FORMAT=json                        # json, text
```

### Loading Configuration

```go
type Config struct {
    Port           int
    DockerNetwork  string
    VelocityImage  string
    MinecraftImage string
    DatabasePath   string
}

func LoadConfig() *Config {
    return &Config{
        Port:           getEnvInt("API_PORT", 8080),
        DockerNetwork:  getEnv("DOCKER_NETWORK", "minecraft-network"),
        VelocityImage:  getEnv("VELOCITY_IMAGE", "itzg/bungeecord:latest"),
        MinecraftImage: getEnv("MINECRAFT_IMAGE", "itzg/minecraft-server:latest"),
        DatabasePath:   getEnv("DATABASE_PATH", "./data/dockermc.db"),
    }
}
```

## Error Handling

### Service Layer

**Pattern:** Log and wrap errors with context:
```go
if err != nil {
    s.logger.ErrorContext(ctx, "Operation failed",
        "server_id", serverID,
        "error", err)
    return fmt.Errorf("failed to do thing: %w", err)
}
```

**Resource Cleanup:**
```go
container, err := s.dockerService.client.ContainerCreate(...)
if err != nil {
    return err
}

// Cleanup on subsequent errors
if err := s.dockerService.client.ContainerStart(...); err != nil {
    s.dockerService.client.ContainerRemove(ctx, container.ID, ...)
    return err
}
```

### Handler Layer

**Pattern:** HTTP status codes with JSON errors:
```go
if err != nil {
    h.logger.ErrorContext(r.Context(), "Failed", "error", err)
    respondError(w, http.StatusInternalServerError, err.Error())
    return
}
```

**Status Codes:**
- 200 OK - Success
- 201 Created - Resource created
- 204 No Content - Successful deletion
- 400 Bad Request - Validation error
- 404 Not Found - Resource not found
- 500 Internal Server Error - Server error

### Graceful Degradation

Non-critical operations log warnings instead of failing:
```go
if err := s.proxyService.ConnectServerToProxy(ctx, server); err != nil {
    s.logger.WarnContext(ctx, "Failed to connect to proxy", "error", err)
    // Continue - server can function standalone
}
```

## Docker Integration

### Container Lifecycle

**Create:**
```go
resp, err := client.ContainerCreate(ctx, &container.Config{
    Image: "itzg/minecraft-server:latest",
    Env: []string{
        "EULA=TRUE",
        "TYPE=PAPER",
        "MAX_PLAYERS=20",
    },
}, &container.HostConfig{
    RestartPolicy: container.RestartPolicy{
        Name: "unless-stopped",
    },
    Mounts: []mount.Mount{
        {
            Type:   mount.TypeVolume,
            Source: volumeID,
            Target: "/data",
        },
    },
}, &network.NetworkingConfig{}, nil, containerName)
```

**Start/Stop:**
```go
client.ContainerStart(ctx, containerID, ...)

timeout := 30  // Graceful shutdown (world save)
client.ContainerStop(ctx, containerID, container.StopOptions{
    Timeout: &timeout,
})
```

**Remove:**
```go
client.ContainerRemove(ctx, containerID, container.RemoveOptions{
    RemoveVolumes: false,  // Keep volumes
    Force:         true,
})
```

### Volume Management

**Create:**
```go
vol, err := client.VolumeCreate(ctx, volume.CreateOptions{
    Name: fmt.Sprintf("mc-server-%s", serverID),
    Labels: map[string]string{
        "server_id": serverID,
        "managed_by": "dockermc",
    },
})
```

**Remove:**
```go
client.VolumeRemove(ctx, volumeID, true)  // Force
```

### Network Management

**Create Bridge Network:**
```go
_, err := client.NetworkCreate(ctx, "minecraft-network", network.CreateOptions{
    Driver: "bridge",
    Labels: map[string]string{
        "managed_by": "dockermc",
    },
})
```

**Connect Container:**
```go
client.NetworkConnect(ctx, "minecraft-network", containerID, &network.EndpointSettings{
    Aliases: []string{"server-name"},  // DNS alias
})
```

### State Inspection

```go
inspect, err := client.ContainerInspect(ctx, containerID)

state := ContainerState{
    Exists:     true,
    Running:    inspect.State.Running,
    Restarting: inspect.State.Restarting,
    Dead:       inspect.State.Dead,
    OOMKilled:  inspect.State.OOMKilled,
}
```

## Database

### GORM Setup

```go
db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Silent),
})

// Auto-migrate schemas
db.AutoMigrate(&models.MinecraftServer{}, &models.ProxyServer{})
```

### Repository Operations

**Create:**
```go
func (r *ServerRepository) Create(server *models.MinecraftServer) error {
    result := r.db.Create(server)
    if result.Error != nil {
        r.logger.Error("Failed to create server", "error", result.Error)
        return result.Error
    }
    return nil
}
```

**Find:**
```go
func (r *ServerRepository) FindByID(id string) (*models.MinecraftServer, error) {
    var server models.MinecraftServer
    result := r.db.First(&server, "id = ?", id)

    if result.Error == gorm.ErrRecordNotFound {
        return nil, fmt.Errorf("server not found")
    }
    if result.Error != nil {
        r.logger.Error("Database error", "error", result.Error)
        return nil, result.Error
    }

    return &server, nil
}
```

**Update:**
```go
func (r *ServerRepository) Update(server *models.MinecraftServer) error {
    result := r.db.Save(server)
    return result.Error
}
```

**Delete:**
```go
func (r *ServerRepository) Delete(id string) error {
    result := r.db.Unscoped().Delete(&models.MinecraftServer{}, "id = ?", id)
    return result.Error
}
```

**Note:** Uses `Unscoped()` for hard deletes (no soft delete).

## CLI Commands

### Serve Command

Start API server:
```bash
go run cmd/api/main.go serve [--port 8080]
```

Flags:
- `--port` - Override API port
- `--log-level` - Set log level (DEBUG, INFO, WARN, ERROR)
- `--log-format` - Set log format (json, text)

### Server Management Commands

```bash
# List servers
go run cmd/api/main.go server list

# Create server
go run cmd/api/main.go server create --name my-server

# Start server
go run cmd/api/main.go server start <id>

# Stop server
go run cmd/api/main.go server stop <id>

# Delete server
go run cmd/api/main.go server delete <id>
```

## Testing

### API Testing

**Bruno Collections:** `api/.bruno/`

Example requests for all endpoints included.

### Manual Testing

```bash
# Health check
curl http://localhost:8080/health

# List servers
curl http://localhost:8080/api/v1/servers

# Create server
curl -X POST http://localhost:8080/api/v1/servers \
  -H "Content-Type: application/json" \
  -d '{"name":"test-server","max_players":20,"motd":"Test Server"}'

# WebSocket logs (use wscat)
wscat -c "ws://localhost:8080/api/v1/servers/{id}/logs?follow=true&tail=100"
```

## Common Development Tasks

### Adding a New Endpoint

1. **Define Model** (if needed) in `internal/models/`
2. **Create Repository** in `internal/database/`
3. **Implement Service** in `internal/service/`
4. **Create Handler** in `internal/api/handlers/`
5. **Add Route** in `internal/api/routes/router.go`
6. **Update OpenAPI** spec in `api/openapi.yaml`

### Adding a New Service Method

```go
func (s *MinecraftServerService) YourMethod(ctx context.Context, req *YourRequest) (*YourResponse, error) {
    s.logger.InfoContext(ctx, "Starting operation", "param", req.Param)

    // Business logic
    result, err := s.doSomething(ctx, req)
    if err != nil {
        s.logger.ErrorContext(ctx, "Operation failed", "error", err)
        return nil, fmt.Errorf("failed to do thing: %w", err)
    }

    s.logger.InfoContext(ctx, "Operation completed", "result", result)
    return result, nil
}
```

**Pattern:**
- Accept `context.Context` first
- Log start, errors, and completion
- Wrap errors with context
- Return wrapped errors

### Adding Middleware

```go
func yourMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Before handler

        next.ServeHTTP(w, r)

        // After handler
    })
}

// In router.go
return yourMiddleware(loggingMiddleware(corsMiddleware(mux)))
```

## Dependencies

### Core Libraries

```go
// Framework
github.com/spf13/cobra v1.10.1           // CLI framework
github.com/joho/godotenv v1.5.1          // .env loading

// Docker
github.com/docker/docker v27.5.1         // Docker SDK
github.com/docker/go-connections v0.6.0  // Docker networking

// Database
gorm.io/gorm v1.31.1                     // ORM
gorm.io/driver/sqlite v1.6.0             // SQLite driver

// HTTP & WebSocket
github.com/coder/websocket v1.8.14       // WebSocket
github.com/swaggo/http-swagger/v2 v2.0.2 // Swagger UI

// Utilities
github.com/google/uuid v1.6.0            // UUID generation
```

## Best Practices

### 1. Always Use Context

```go
func (s *Service) Method(ctx context.Context, ...) error {
    // Pass context to all downstream calls
    result, err := s.docker.client.ContainerCreate(ctx, ...)
}
```

### 2. Structured Logging

```go
s.logger.InfoContext(ctx, "Message",
    "key1", value1,
    "key2", value2,
)
```

### 3. Error Wrapping

```go
if err != nil {
    return fmt.Errorf("failed to create container: %w", err)
}
```

### 4. Resource Cleanup

```go
defer conn.Close()
defer logReader.Close()

// Cleanup on error
if err != nil {
    client.ContainerRemove(ctx, containerID, ...)
    return err
}
```

### 5. Graceful Shutdown

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

server.Shutdown(ctx)
```

### 6. Idempotent Operations

```go
// Check if already exists before creating
if exists {
    return nil
}

// Create
```

## Known Limitations

1. **Single Proxy** - Only supports one proxy instance
2. **No Authentication** - API is unauthenticated
3. **No Rate Limiting** - Unprotected against abuse
4. **No TLS** - HTTP only (no HTTPS)
5. **SQLite Only** - No PostgreSQL/MySQL support
6. **No Backups** - Volume backups not automated
7. **Local Docker Only** - No remote Docker support

## Future Enhancements

- [ ] Multi-proxy support
- [ ] Authentication & authorization
- [ ] Rate limiting
- [ ] TLS/HTTPS support
- [ ] PostgreSQL support
- [ ] Automated backups
- [ ] Remote Docker daemon support
- [ ] Metrics and monitoring
- [ ] Resource limits (CPU/memory)
- [ ] Health checks for containers

## Troubleshooting

### Database Issues

```bash
# Check database
sqlite3 ./data/dockermc.db ".tables"
sqlite3 ./data/dockermc.db "SELECT * FROM minecraft_servers;"
```

### Docker Issues

```bash
# Check Docker daemon
docker ps
docker network ls
docker volume ls

# Check logs
docker logs mc-server-{id}
docker logs mc-proxy-main
```

### Log Analysis

```bash
# View application logs (JSON format)
LOG_FORMAT=json go run cmd/api/main.go serve

# Pretty print JSON logs
go run cmd/api/main.go serve 2>&1 | jq
```

## Related Files

- **Frontend:** `dockermc-cloud-manager-ui/CLAUDE.md`
- **OpenAPI Spec:** `api/openapi.yaml`
- **API Tests:** `api/.bruno/`
- **Environment:** `.env`

## Key Files Reference

- `cmd/api/main.go` - Application entry point
- `internal/cli/commands/serve.go` - Server initialization
- `internal/api/routes/router.go` - HTTP routing
- `internal/service/minecraft.go` - Server business logic
- `internal/service/proxy.go` - Proxy business logic
- `internal/service/docker.go` - Docker integration
- `internal/database/database.go` - Database layer
- `internal/models/` - Data models
