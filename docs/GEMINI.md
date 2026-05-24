# Gemini Codebase Structural Changes Log

This document tracks structural changes made to the repository code.

## 2026-05-22: AI Service Improvements and Backend Health Monitoring

### Python AI Service (`/ai-service`)
- **[Dependency Update]**: Added `pika>=1.3.0` to [requirements.txt](file:///Users/rubenalves/Documents/repos/_school/iot/final/ai-service/requirements.txt) to support RabbitMQ messaging queue ingestion.
- **[Model Helpers]**: Added dataset generation, normalization, loading, and training helper functions to [model.py](file:///Users/rubenalves/Documents/repos/_school/iot/final/ai-service/src/model.py).
- **[Refactoring & Features]**: Refactored [main.py](file:///Users/rubenalves/Documents/repos/_school/iot/final/ai-service/src/main.py):
  - Created a baseline model on startup if not present to prevent service crashes.
  - Implemented the gRPC `PredictSeverity` endpoint to parse features (`fails`, `distance_cm`, `is_denied`) from real `SensorEvent` protobuf structures.
  - Implemented the gRPC `RetrainModel` endpoint to retrain the neural network on arbitrary dataset paths (CSV) and reload it dynamically in-memory with thread safety.
  - Implemented an asynchronous background daemon thread (`RabbitMQConsumer`) consuming events from the `sensor_events` queue and recording them into `data/sensor_events.csv` for self-supervised data compilation.

### Go Backend (`/backend`)
- **[NEW]** Created [backend/internal/config/config.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/config/config.go) to implement configuration loading with `caarlos0/env`.
- **[NEW]** Created [backend/internal/monitor/monitor.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/monitor/monitor.go) to run the background service health monitoring.
- **[NEW]** Created [backend/cmd/api/main.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/cmd/api/main.go) as the new API service entrypoint using `go-chi`. It includes endpoints for health status and AI retraining.
- **[DELETE]** Deleted `backend/cmd/main-server/main.go` and the `main-server` directory.
- **[Dependency Update]**: Added `github.com/go-chi/chi/v5` for REST API routing, `github.com/caarlos0/env/v11` for configuration, `github.com/lib/pq` for PostgreSQL, and `github.com/influxdata/influxdb-client-go/v2` for InfluxDB to `go.mod`.
- **[Broker Refactoring]**:
  - Modified [backend/broker/mqtt.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/broker/mqtt.go): Changed `StartSubscriber` to return `(mqtt.Client, error)` and enabled auto-reconnect.
  - Modified [backend/broker/rabbitmq.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/broker/rabbitmq.go): Added sync.RWMutex, connection checking, connection recovery, and graceful close.
- **[Deployment configuration]**:
  - Modified [backend/Dockerfile](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/Dockerfile): Updated build directive to compile `./cmd/api/main.go` and run `./api-server`.
  - Modified [deployments/docker-compose.yml](file:///Users/rubenalves/Documents/repos/_school/iot/final/deployments/docker-compose.yml): Added static admin token environment variable to InfluxDB, set database URLs on main-server, and updated dependencies.

## 2026-05-22: Simplified DDD Refactor and Core Subfolder

### Firmware (`/firmware`)
- **[Refactoring & Features]**: Modified [main.cpp](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/refactor-backend-domain-structure/firmware/src/main.cpp) to transmit the `rfid_uid` field in authentication telemetry payloads (e.g., `access_granted`, `access_denied`).

### Go Backend (`/backend`)
- **[NEW]** Created `internal/core` package consolidating all shared infrastructure clients, configurations, health check Pings, and close cleanups:
  - [postgres.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/refactor-backend-domain-structure/backend/internal/core/postgres.go): Schema migration for `users` and `telemetry` tables and client setup.
  - [mqtt.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/refactor-backend-domain-structure/backend/internal/core/mqtt.go): Paho MQTT subscriber setup.
  - [influx.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/refactor-backend-domain-structure/backend/internal/core/influx.go): InfluxDB v2 integration with WriteAPI.
  - [rabbitmq.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/refactor-backend-domain-structure/backend/internal/core/rabbitmq.go): Queue publishers, reconnect handlers, and check states.
- **[NEW]** Created `user` domain layer managing RFID cards/tags and appending personal data:
  - [entity.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/refactor-backend-domain-structure/backend/internal/domain/user/entity.go): Domain structure definition.
  - [repository.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/refactor-backend-domain-structure/backend/internal/domain/user/repository.go): Postgres user storage interface and implementation.
  - [service.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/refactor-backend-domain-structure/backend/internal/domain/user/service.go): Register and update user domain business logic.
  - [handler.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/refactor-backend-domain-structure/backend/internal/domain/user/handler.go): HTTP handler for user query (`GET /api/users`) and details update (`PUT /api/users/{uid}`).
- **[NEW]** Created `telemetry` domain layer for ingestion pipeline and audit logs:
  - [repository.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/refactor-backend-domain-structure/backend/internal/domain/telemetry/repository.go): PostgreSQL telemetry logs persistence.
  - [service.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/refactor-backend-domain-structure/backend/internal/domain/telemetry/service.go): Ingestion orchestrator that persists telemetry, notifies RabbitMQ, triggers AI security evaluation, and automatically registers new RFID cards in the user domain.
  - [handler.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/refactor-backend-domain-structure/backend/internal/domain/telemetry/handler.go): Ingestion simulation endpoint `POST /api/telemetry` for system manual verification.
- **[NEW]** Created `ai` domain adapter pattern decoupled from the gRPC definition:
  - [service.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/refactor-backend-domain-structure/backend/internal/domain/ai/service.go): Clean AI interface for Go backend ports.
  - [grpc_client.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/refactor-backend-domain-structure/backend/internal/domain/ai/grpc_client.go): Adapt client to the gRPC client endpoints.
- **[DELETE]** Removed the old broker package (`backend/broker/mqtt.go`, `backend/broker/rabbitmq.go`).
- **[Refactoring & Features]**:
  - Updated [models.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/refactor-backend-domain-structure/backend/internal/models/models.go) to support the optional `rfid_uid` field in sensor payload.
  - Updated [monitor.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/refactor-backend-domain-structure/backend/internal/monitor/monitor.go) to utilize the unified `internal/core` components for checking connections.
  - Updated [main.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/refactor-backend-domain-structure/backend/cmd/api/main.go) to initialize the DDD domains and route API calls.
- **[Tests]**: Added new unit tests verifying service logic in users and telemetry domains:
  - [user/service_test.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/refactor-backend-domain-structure/backend/internal/domain/user/service_test.go)
  - [telemetry/service_test.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/refactor-backend-domain-structure/backend/internal/domain/telemetry/service_test.go)

## 2026-05-24: Porting Collaborator Web Server and Backend Changes

### Firmware (`/firmware`)
- **[Aesthetics & Features]**: Modified [index.h](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/integrate-web-server-features/firmware/include/web/index.h) and [wifi.h](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/integrate-web-server-features/firmware/include/web/wifi.h) to replace default layouts with a premium dark-themed glassmorphism interface. Included AJAX controllers for unlocking, live status polling, consolidated health checking, and a registered users explorer.
- **[Refactoring & Features]**: Refactored [main.cpp](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/integrate-web-server-features/firmware/src/main.cpp):
  - Added an auto-close timer (`autoCloseTimer`) using the `Timer` library to automatically relock the door 5 seconds after opening.
  - Set up dynamic WiFi credentials loading from ESP32 `Preferences` (NVRAM) and saving via `/wifi-save` (triggers automatic reboot).
  - Implemented `/open`, `/status`, `/wifi-info`, `/users`, `/user-details`, and `/check-services` endpoints, acting as a JSON/REST proxy to the Go backend.

### Go Backend (`/backend`)
- **[Routing Update]**: Registered the `GET /api/users/{uid}` endpoint in [handler.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/integrate-web-server-features/backend/internal/domain/user/handler.go) to fetch detailed profile information for specific card UIDs.
- **[MQ Restructuring]**: Modified [rabbitmq.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/integrate-web-server-features/backend/internal/core/rabbitmq.go) to declare a new `heartbeat_events` queue and implement methods for publishing and consuming heartbeat events.
- **[Refactoring & Pipeline]**:
  - Updated `Ingest()` in [service.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/integrate-web-server-features/backend/internal/domain/telemetry/service.go) to publish heartbeat messages to RabbitMQ and return immediately, bypassing synchronous database saves and AI service evaluations.
  - Updated [main.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/integrate-web-server-features/backend/cmd/api/main.go) to launch an asynchronous heartbeat queue consumer that writes heartbeat events to the database in the background.
- **[Tests]**: Added `TestTelemetryIngestHeartbeat` to [service_test.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/integrate-web-server-features/backend/internal/domain/telemetry/service_test.go) to validate the new heartbeat offloading flow.


