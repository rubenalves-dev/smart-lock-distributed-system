# structural Changes Log (GEMINI.md)

This document tracks structural changes made to the repository code.

## 2026-05-22: Backend Refactoring and Health Monitoring Implementation

### Directory Structure & File Additions/Deletions
- **[NEW]** Created [docs/GEMINI.md](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/backend-health-monitor-services/docs/GEMINI.md) (this file) to track codebase changes.
- **[NEW]** Created [backend/internal/config/config.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/backend-health-monitor-services/backend/internal/config/config.go) to implement configuration loading with `caarlos0/env`.
- **[NEW]** Created [backend/internal/monitor/monitor.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/backend-health-monitor-services/backend/internal/monitor/monitor.go) to run the background service health monitoring.
- **[NEW]** Created [backend/cmd/api/main.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/backend-health-monitor-services/backend/cmd/api/main.go) as the new API service entrypoint using `go-chi`.
- **[DELETE]** Deleted `backend/cmd/main-server/main.go` and the `main-server` directory.

### Code & Library Updates
- **Dependencies (`backend/go.mod`)**:
  - Added `github.com/go-chi/chi/v5` for REST API routing.
  - Added `github.com/caarlos0/env/v11` for configuration.
  - Added `github.com/lib/pq` as the PostgreSQL database driver.
  - Added `github.com/influxdata/influxdb-client-go/v2` as the InfluxDB client.
- **Broker Refactoring**:
  - Modified [backend/broker/mqtt.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/backend-health-monitor-services/backend/broker/mqtt.go): Changed `StartSubscriber` to return `(mqtt.Client, error)` to make the client reference available to the health checker. Enabled auto-reconnect.
  - Modified [backend/broker/rabbitmq.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/backend-health-monitor-services/backend/broker/rabbitmq.go): Added sync.RWMutex, connection checking (`IsConnected() bool`), connection recovery (`Reconnect(url string) error`), and a graceful `Close()` method to handle network interruptions.
- **Deployment configuration**:
  - Modified [backend/Dockerfile](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/backend-health-monitor-services/backend/Dockerfile): Updated build directive to compile `./cmd/api/main.go` and run `./api-server`.
  - Modified [deployments/docker-compose.yml](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/backend-health-monitor-services/deployments/docker-compose.yml): Added static admin token environment variable `DOCKER_INFLUXDB_INIT_ADMIN_TOKEN` to InfluxDB initialization, set database URLs/configuration on main-server, and updated dependencies in `depends_on`.
