# Gemini Codebase Structural Changes Log

This document tracks structural changes made to the repository code.

## 2026-06-01: Continuous RFID Polling and Reader Diagnostics

### ESP32 Firmware (`/arduino-ide` & `/firmware`)
- **[Continuous RFID Check & Rate-Limiting]**: Modified [main.ino](file:///Users/rubenalves/Documents/repos/_school/iot/final/arduino-ide/main/main.ino) and [main.cpp](file:///Users/rubenalves/Documents/repos/_school/iot/final/firmware/src/main.cpp) to remove the proximity/ultrasonic gating (`isClose` / `isObjectClose`) on the RFID sensor. Implemented a non-blocking 100ms polling rate-limiter using `millis()` to allow the MFRC522 RF antenna field to charge passive cards stably, resolving scan failures.
- **[Reader State Diagnostics]**: Modified `RFID.h` and `RFID.cpp` in both the [arduino-ide](file:///Users/rubenalves/Documents/repos/_school/iot/final/arduino-ide/main/RFID.h) and [firmware](file:///Users/rubenalves/Documents/repos/_school/iot/final/firmware/lib/RFID/RFID.h) directories to implement `getVersion()` (queries MFRC522 `VersionReg`) and `isConnected()` (verifies connection status) helper methods.
- **[RFID Check Bug Fix]**: Fixed a buffer over-read bug in the `check()` method inside [RFID.cpp](file:///Users/rubenalves/Documents/repos/_school/iot/final/arduino-ide/main/RFID.cpp) by ensuring the UID size matches 4 bytes and restricting the `memcmp` operation size to 4 bytes, avoiding undefined behavior when 7-byte or 10-byte tags are scanned.
- **[Diagnostic Logging]**:
  - Updated RFID initialization setup methods to print the specific card reader chip software version to Serial (identifying versions `v1.0`, `v2.0`, `clone`, or `WARNING` connection failures).
  - Updated the periodic status print routines to include the RFID connection state (`CONNECTED` with version or `DISCONNECTED`).
  - Formatted the MQTT subscription callback logs to print the target topic and details when `UNLOCK` and `LOCK` commands are received.

## 2026-06-01: Major Overhaul of Model Evaluation, Retraining, History Storage, User Activation, Heartbeat Storing, Device Control and Uptime Charts

### Database Schema (`/backend`)
- **[NEW] [AI History Tables]**: Modified [postgres.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/core/postgres.go) to declare and run migrations for `ai_evaluations` and `ai_retrains` history tracking tables on startup.

### Python AI Service (`/ai-service`)
- **[Evaluation CSV Parsing Fix]**: Modified `EvaluateModel` in [main.py](file:///Users/rubenalves/Documents/repos/_school/iot/final/ai-service/src/main.py) to reliably detect inline CSV contents vs file paths, resolving file evaluation crashes.
- **[Retraining Inline CSV Support]**: Updated `RetrainModel` in [main.py](file:///Users/rubenalves/Documents/repos/_school/iot/final/ai-service/src/main.py) to write inline CSV payloads (such as database telemetry data dumps) to a temporary file, retrain, and clean it up automatically.

### Go Backend (`/backend`)
- **[Telemetry Heartbeat Storage]**: Modified `Ingest` in [service.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/domain/telemetry/service.go) to store periodic heartbeat payloads in PostgreSQL, ensuring ESP32 heartbeats are tracked in the database.
- **[Telemetry List & Devices Queries]**: Added `GetAll` and `GetDevices` queries to [repository.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/domain/telemetry/repository.go) and [service.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/domain/telemetry/service.go) to select all telemetry logs (and generate training CSVs) and find active device IDs.
- **[AI Database Persistence]**: Added database operations to `SaveRetrain`, `ListRetrains`, and `ListEvaluations` in [repository.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/domain/ai/repository.go) and [service.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/domain/ai/service.go), mapping evaluations/retrains logs to PostgreSQL and sanitizing dataset path string column lengths.
- **[New REST Endpoints]**: Registered `GET /api/ai/evaluations`, `GET /api/ai/retrains`, and `GET /api/telemetry/devices` routes in [main.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/cmd/api/main.go) and updated the `POST /api/ai/retrain` endpoint to dump telemetry logs as CSV if requested with `"database"`.
- **[Decoupled Client Interfaces]**: Adjusted signature of `NewGRPCClient` in [grpc_client.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/domain/ai/grpc_client.go) to return the decoupled `GRPCClient` interface. Updated tests mocks in [service_test.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/domain/telemetry/service_test.go).

### Vue Frontend (`/frontend`)
- **[User Activation PUT Fix]**: Modified [UsersView.vue](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/views/Tables/UsersView.vue) to explicitly send `is_accepted: true` when registering a pending RFID card, enabling the AI verification pipeline on subsequent swipes.
- **[Uptime Charts Query Interval]**: Modified [UptimeChart.vue](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/components/charts/LineChart/UptimeChart.vue) to query pings dynamically through `API_BASE_URL` with 5-minute intervals inside a 24-hour range.
- **[Device Selection Controls]**: Added a dynamic device selection dropdown at the top of [Device.vue](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/components/device/Device.vue), querying active device IDs from the backend and loading corresponding parameters.
- **[Model Evaluation View]**: Integrated evaluations database history querying on mount in [AiEvaluationView.vue](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/views/Pages/AiEvaluationView.vue) and added a header format note.
- **[AI Retraining View]**: Refactored [AiRetrainView.vue](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/views/Pages/AiRetrainView.vue) to include a mode switcher (Base Dados, Ficheiro Local, Caminho Servidor), local file drag-and-drop parser with a header format note, and history query synchronization on mount.

## 2026-05-31: Multi-Factor Authentication (MFA) and Real-time Notifications

### Database Schema (`/backend`)
- **[NEW] [MFA Table Migration]**: Modified [postgres.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/mfa-door-notification-ui/backend/internal/core/postgres.go) to initialize the `mfa_requests` table on startup, tracking card scans flagged by the AI service.

### Go Backend (`/backend`)
- **[NEW] [WebSocket Hub]**: Created [websocket.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/mfa-door-notification-ui/backend/internal/core/websocket.go) implementing connection management and JSON payload broadcasting to frontend clients.
- **[NEW] [MFA Domain Components]**: Created `internal/domain/mfa` package to define:
  - [entity.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/mfa-door-notification-ui/backend/internal/domain/mfa/entity.go): Request data structure.
  - [repository.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/mfa-door-notification-ui/backend/internal/domain/mfa/repository.go): PostgreSQL Create, ListAll, and UpdateStatus commands.
  - [service.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/mfa-door-notification-ui/backend/internal/domain/mfa/service.go): Approval (MQTT unlock broadcast) and rejection (permanent card block in user domain) business flows.
  - [handler.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/mfa-door-notification-ui/backend/internal/domain/mfa/handler.go): HTTP controllers for lists, approval, and rejection.
- **[Lock Command Bugfix]**: Modified [mqtt.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/mfa-door-notification-ui/backend/internal/core/mqtt.go) to publish `"UNLOCK"` commands to topic `"lock/commands"`, resolving a mismatch where the backend was publishing to an unmonitored topic.
- **[Telemetry Ingest Pipeline]**: Refactored `Ingest` in [service.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/mfa-door-notification-ui/backend/internal/domain/telemetry/service.go) to handle `"access_request"` online check events. Suspicious/critical attempts (AI classification >= 2) hold the lock and trigger MFA request creation; normal attempts (classification < 2) unlock immediately.
- **[Routing Registration]**: Updated [main.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/mfa-door-notification-ui/backend/cmd/api/main.go) to map websocket connection upgrades to `/api/ws` and register mfa routing blocks.
- **[Unit Tests Verification]**: Modified mock integrations and appended `TestTelemetryIngestRFIDAIMFA` in [service_test.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/mfa-door-notification-ui/backend/internal/domain/telemetry/service_test.go) to assert telemetry-to-MFA transitions.

### Vue Frontend (`/frontend`)
- **[NEW] [MFA Composable]**: Created [useMfa.ts](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/mfa-door-notification-ui/frontend/src/composables/useMfa.ts) to manage state, fetch requests, call approve/reject endpoints, and handle real-time updates over WebSocket streams.
- **[NEW] [MFA Approvals View]**: Created [MfaRequestsView.vue](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/mfa-door-notification-ui/frontend/src/views/Pages/MfaRequestsView.vue) displaying pending access requests, sensor details, AI classification confidence levels, and quick approve/block actions. Built-in with an interactive **Telemetry Simulator** panel allowing administrators to simulate RFID swipes with custom distances, fail counts, and light levels directly from the dashboard view for testing.
- **[Sidebar Navigation]**: Added "Aprovações MFA" link to [AppSidebar.vue](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/mfa-door-notification-ui/frontend/src/components/layout/AppSidebar.vue).
- **[Notification Dropdown Alerts]**: Integrated live alerts into [NotificationMenu.vue](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/mfa-door-notification-ui/frontend/src/components/layout/header/NotificationMenu.vue) to display dynamic indicators and click-to-redirect shortcuts when new MFA requests are broadcast.
- **[Router Config]**: Registered path `/mfa/requests` in [index.ts](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/mfa-door-notification-ui/frontend/src/router/index.ts).

### ESP32 Firmware (`/arduino-ide`)
- **[Remote Auth Delegation]**: Modified [main.ino](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/mfa-door-notification-ui/arduino-ide/main/main.ino) to change RFID check behavior. When online, accepted card scans publish an `access_request` telemetry payload and await the backend MQTT command instead of unlocking locally. Offline cache lookup fallbacks remain active.

### Development Configuration (`/deployments`, `/`)
- **[Hot-reloading watch target]**: Added `develop.watch` configuration for `ai-service`, `main-server`, and `frontend` services to [docker-compose.yml](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/mfa-door-notification-ui/deployments/docker-compose.yml), and defined the `make watch` target in [Makefile](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/mfa-door-notification-ui/Makefile) to automatically trigger hot-rebuilding of modified containers on host file changes.

## 2026-05-31: AI Model Retraining Page and Diagnostics

### Python AI Service (`/ai-service`)
- **[Diagnostics Response]**: Modified `RetrainModel` method in [main.py](file:///Users/rubenalves/Documents/repos/_school/iot/final/ai-service/src/main.py) to extract training history losses (`loss` and `val_loss`) and populate the gRPC `TrainingDiagnostics` message inside `RetrainModelResponse`.

### Go Backend (`/backend`)
- **[Model Structures]**: Added `TrainingDiagnostics` and `RetrainResult` data structures to [models.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/models/models.go).
- **[AI Client Integration]**: Updated `RetrainModel` interface definition in [interfaces.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/domain/ai/interfaces.go), client implementation in [grpc_client.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/domain/ai/grpc_client.go), and domain delegation in [service.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/domain/ai/service.go) to return the parsed training diagnostics results.
- **[Mocks Update]**: Modified mock `fakeAIService` definition in [service_test.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/domain/telemetry/service_test.go) to satisfy the new interface.
- **[Web Routing Setup]**: Registered HTTP endpoint `POST /api/ai/retrain` inside [main.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/cmd/api/main.go) for triggering models retraining.

### Vue Frontend (`/frontend`)
- **[Retrain Composable]**: Created [useAiRetrain.ts](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/composables/useAiRetrain.ts) helper composable wrapping API fetches to the backend retrain router route.
- **[Aesthetics & UI Pages]**: Created [AiRetrainView.vue](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/views/Pages/AiRetrainView.vue) view page displaying epochs/dataset input controls, visual accuracy/loss progress meters, fitting analysis warning alerts (overfitting vs underfitting vs balanced model), and a session retraining log table.
- **[Router Config]**: Registered the `/ai/retrain` SPA view route inside [index.ts](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/router/index.ts).
- **[Sidebar Navigation]**: Added "AI Retraining" link config entry in [AppSidebar.vue](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/components/layout/AppSidebar.vue).

## 2026-05-24: Backend API and Integration Surface Updates

### API Contract (`/api`)
- **[Protobuf Extension]**: Extended [lock.proto](file:///Users/rubenalves/Documents/repos/_school/iot/final/api/proto/lock.proto) to include `EvaluateModel` RPC method under `AIService` along with `EvaluateModelRequest`, `EvaluateModelResponse`, `ConfusionMatrixRow`, `EvaluationMetrics`, and `BinaryEvaluationMetrics` messages.

### Go Backend (`/backend`)
- **[Model Definitions]**: Added data structures for AI evaluation results and InfluxDB time-series series health to [models.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/models/models.go).
- **[InfluxDB Integration]**: Added `QueryHealth` method to [influx.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/core/influx.go) to run parameterized Flux queries for service uptime.
- **[AI Domain Client]**: Updated [service.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/domain/ai/service.go) and [grpc_client.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/domain/ai/grpc_client.go) to declare and implement the gRPC client wrapper for `EvaluateModel`.
- **[Users Filtering]**: Added support for filtering incomplete profiles (`?incomplete=true`) in user list handler in [handler.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/domain/user/handler.go).
- **[Telemetry and Lock Control]**:
  - Implemented `GetLatest` db query in [repository.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/domain/telemetry/repository.go).
  - Implemented `GetLatestTelemetry` and `UnlockDoor` methods in [service.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/domain/telemetry/service.go).
  - Added REST routes `GET /api/telemetry/latest` and `POST /api/door/unlock` in [handler.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/domain/telemetry/handler.go).
- **[Web Routing Setup]**: Registered new HTTP endpoints `/api/metrics/health` and `/api/ai/evaluate` inside [main.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/cmd/api/main.go).
- **[Mocks Update]**: Updated unit test mocks in [service_test.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/domain/telemetry/service_test.go) to satisfy the extended interfaces.

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
- **[Tests]**: Added `TestTelemetryIngestHeartbeat` to [service_test.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/integrate-web-server-features/backend/internal/domain/telemetry/service_test.go) to validate the new heartbeat offloading flow.

## 2026-05-26: Frontend Build Fixes

### Vue Frontend (`/frontend`)
- **[Aesthetics & Styles]**: Fixed CSS syntax warnings in [main.css](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/assets/main.css) by removing Tailwind v4 `dark:bg-gray-700` `@apply` from the nested `&::-webkit-scrollbar-thumb` pseudo-element selector and defining it explicitly for both `.custom-scrollbar` and `.fc-view-harness` classes.
- **[Refactoring & Compilation]**:
  - Removed unused variable declarations (`dropdownOpen`, `notifying`) and the `toggleDropdown` method in [AppHeader.vue](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/components/layout/AppHeader.vue).
  - Removed unused `computed` import and the unused `const props =` assignment in [Alert.vue](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/components/ui/Alert.vue).
  - Removed unused `const props =` assignment in [Avatar.vue](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/components/ui/Avatar.vue).
  - Removed unused `computed` import in [Button.vue](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/components/ui/Button.vue).
  - Prefixed unused `to` and `from` route parameters with underscores in [index.ts](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/router/index.ts) to satisfy TypeScript `noUnusedLocals` linting.

## 2026-05-27: Fixed ESPAsyncWebServer / AsyncTCP const compatibility in Arduino IDE

### Arduino IDE Vendored Libraries (`/arduino-ide/libraries`)
- **[Const Correctness Patch]**: Modified [ESPAsyncWebServer.h](file:///Users/rubenalves/Documents/repos/_school/iot/final/arduino-ide/libraries/ESP_Async_WebServer/src/ESPAsyncWebServer.h) to cast away constness from `this` when calling `_server.status()` inside `state() const`. This fixes the `passing 'const AsyncServer' as 'this' argument discards qualifiers` compilation error when building the Arduino sketch with ESP32 support.

## 2026-05-28: Ported web server and firmware logic to Arduino IDE with built-in WebServer

### Arduino IDE Sketch (`/arduino-ide/main`)
- **[Refactoring & Compilation]**: Refactored [main.ino](file:///Users/rubenalves/Documents/repos/_school/iot/final/arduino-ide/main/main.ino) to implement all the logic from the firmware's `main.cpp`. Replaced `ESPAsyncWebServer` with the standard, modern, built-in synchronous `WebServer.h` library (and explicitly included `WiFi.h`) to ensure compatibility with modern ESP32 cores and compiling out of the box in the Arduino IDE. Restructured the main `loop()` to run `server.handleClient()`, `mqtt.update()`, and timer updates in a non-blocking fashion while checking sensors on a 1-second interval using `millis()`.
- **[Renaming File]**: Renamed `WIFI.h` to [WIFI_HTML.h](file:///Users/rubenalves/Documents/repos/_school/iot/final/arduino-ide/main/WIFI_HTML.h) to resolve case-insensitivity file name collision on macOS/Windows where `#include <WiFi.h>` conflicts with the local header.
- **[WiFi Access Point Fallback]**: Added an Access Point (softAP) mode fallback in [main.ino](file:///Users/rubenalves/Documents/repos/_school/iot/final/arduino-ide/main/main.ino) that runs if connection to the local network fails. It exposes a SSID named `SmartLock-Setup` (password: `12345678`) hosting the configuration page at `http://192.168.4.1`. Modified `/wifi-info` to return the AP IP if active.
- **[Fix Boot Crash]**: Moved MQTT initialization (`mqtt.setup()`) to execute *after* WiFi initialization (`WiFi.begin()` or `WiFi.softAP()`) in [main.ino](file:///Users/rubenalves/Documents/repos/_school/iot/final/arduino-ide/main/main.ino) to prevent a FreeRTOS semaphore crash on boot caused by trying to run socket/TCP routines before the network interface is initialized.
- **[Optimize Synchronous Performance]**: Added WiFi connectivity checks around `/users`, `/user-details`, `/check-services`, `sendTelemetry()`, and `mqtt.update()` inside [main.ino](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/rfid-access-control-system/arduino-ide/main/main.ino). This prevents the synchronous web server thread from blocking on outbound network timeouts and MQTT reconnection attempts when the ESP32 is running in Access Point mode (or disconnected from WiFi), allowing pages at `http://192.168.4.1` to load instantly.

## 2026-05-29: Dynamically Stored Heartbeats, Access Control Validation, and Premium Users Management View

### Database Schema (`/backend`)
- **[PostgreSQL Schema Alteration]**: Modified [postgres.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/rfid-access-control-system/backend/internal/core/postgres.go) to run `ALTER TABLE users ADD COLUMN IF NOT EXISTS is_accepted BOOLEAN DEFAULT FALSE;` and `ADD COLUMN IF NOT EXISTS is_blocked BOOLEAN DEFAULT FALSE;` at startup to support card activation and blocking states.

### Go Backend (`/backend`)
- **[User Entity & Repository]**: Modified [entity.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/rfid-access-control-system/backend/internal/domain/user/entity.go) and [repository.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/rfid-access-control-system/backend/internal/domain/user/repository.go) to support reading, inserting, and updating the new `is_accepted` and `is_blocked` fields.
- **[User Service & Handler]**: Modified [service.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/rfid-access-control-system/backend/internal/domain/user/service.go) and [handler.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/rfid-access-control-system/backend/internal/domain/user/handler.go) to expose the new fields in API requests (`PUT /api/users/{uid}`) and pass them cleanly.
- **[Telemetry/RFID Ingestion]**: Refactored `Ingest()` in [service.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/rfid-access-control-system/backend/internal/domain/telemetry/service.go):
  - Automatically records heartbeats in the database via the background consumer (and persists them even if RabbitMQ is bypassed).
  - Intercepts all telemetry events containing a scanned RFID card's UID.
  - **First Scan Auto-Registration**: If the card UID does not exist in the database, registers it automatically as pending (not accepted) and logs the event under status `"Pending"`.
  - **Subsequent Scan Access Check**: If it's not the first scan, verifies if the card is active and not blocked. If active, triggers the MQTT unlock command and logs `access_granted`. Otherwise, restricts access and logs `access_denied` indicating whether it's pending or blocked.
- **[Tests]**: Updated mock structures in [service_test.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/rfid-access-control-system/backend/internal/domain/user/service_test.go) and added `TestTelemetryIngestRFIDAccessControl` in [service_test.go](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/rfid-access-control-system/backend/internal/domain/telemetry/service_test.go) to fully assert dynamic registration, activation, blocking, and door unlock actions.

### Vue Frontend (`/frontend`)
- **[Aesthetics & Features]**: Rewrote [UsersView.vue](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/rfid-access-control-system/frontend/src/views/Tables/UsersView.vue) to wrap the list of users inside premium layout components (`<AdminLayout>`, `<PageBreadcrumb>`, and `<ComponentCard>`) and styled the table with Tailwind CSS dark/light modes.
- **[Actions & Modals]**: Integrated edit forms within a modal dialog to let the administrator input names and emails to accept and activate pending cards, or modify existing active users, along with quick buttons to block/unblock access.
- **[SPA Refresh Fix]**: Updated the command in [Dockerfile](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/rfid-access-control-system/frontend/Dockerfile) to serve the built static assets with the SPA flag (`-s` / `--spa`), resolving 404 page-not-found errors upon browser refreshing.

### Arduino IDE Sketch (`/arduino-ide`)
- **[RFID SPI Fix]**: Modified [RFID.cpp](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/rfid-access-control-system/arduino-ide/main/RFID.cpp) to pass `sdaPin_` as the fourth parameter to `SPI.begin()`. This configures the ESP32 SPI subsystem to control the MFRC522 card reader's Chip Select line properly.
- **[Dynamic RFID & Preferences Cache]**:
  - Declared and implemented the public method `readCard()` in [RFID.h](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/rfid-access-control-system/arduino-ide/main/RFID.h) and [RFID.cpp](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/rfid-access-control-system/arduino-ide/main/RFID.cpp) to read cards and populate buffer variables directly without hardcoded matches.
  - Rewrote the card scanning loop in [main.ino](file:///Users/rubenalves/.gemini/antigravity/worktrees/final/rfid-access-control-system/arduino-ide/main/main.ino) to implement dynamic HTTP query calls. During active connection, it requests card state (`is_accepted` / `is_blocked`) from `/api/users/{uid}` and stores it locally under ESP32 `Preferences` namespace `"cards"`.
  - Implemented the offline fallback: if internet connection fails or server is unreachable, the system queries the local `Preferences` cards cache to grant access (matches active state) or deny access (pending/blocked).

## 2026-05-29: Optimized RFID Scanning and Decoupled Polling in Arduino IDE

### Arduino IDE Sketch (`/arduino-ide`)
- **[RFID Polling Optimization]**: Modified [main.ino](file:///Users/rubenalves/Documents/repos/_school/iot/final/arduino-ide/main/main.ino) to move the card reading logic (`rfid.readCard()`) out of the slow 1-second sensor update block. The RFID reader is now polled continuously on every loop cycle, ensuring instantaneous swipe detection.
- **[Gating Optimization]**: Gated the continuous RFID polling by `ultrassonic.isObjectClose()` (which checks the cached distance state). This preserves the proximity requirement without calling blocking ultrasonic routines on every loop iteration.
- **[Silenced Verbose Logs]**: Modified [RFID.cpp](file:///Users/rubenalves/Documents/repos/_school/iot/final/arduino-ide/main/RFID.cpp) to remove the `"No card present"` and `"No card present or failed to read the card!"` print statements, avoiding console flooding when the card reader is queried rapidly.

## 2026-05-29: Added Serial Diagnostics to RFID Reading and Telemetry Ingestion

### Arduino IDE Sketch (`/arduino-ide`)
- **[Diagnostics & Features]**: Modified [main.ino](file:///Users/rubenalves/Documents/repos/_school/iot/final/arduino-ide/main/main.ino) to add comprehensive serial print logs when an RFID card is scanned:
  - Prints the target URL and status/response of the backend API requests during online card checks.
  - Logs local preferences cache reads, writes, and offline fallbacks.
  - Logs generated JSON telemetry payloads and MQTT publications.
  - Logs lock state transitions and the triggering user/reason.

## 2026-05-29: Migrated to Docker Compose v2 in CI/CD Workflow and Makefile

### CI/CD Deployment Configuration (`/.github/workflows`, `/`)
- **[Docker Compose Integration]**: Modified [.github/workflows/main.yml](file:///Users/rubenalves/Documents/repos/_school/iot/final/.github/workflows/main.yml) and [Makefile](file:///Users/rubenalves/Documents/repos/_school/iot/final/Makefile) to replace legacy `docker-compose` (hyphenated v1 command) with modern `docker compose` (v2 plugin).
  - Added a self-contained fallback installer block in [.github/workflows/main.yml](file:///Users/rubenalves/Documents/repos/_school/iot/final/.github/workflows/main.yml): If `docker compose` is not globally installed or active as a Docker CLI plugin on the VPS, it auto-detects architecture (`x86_64` vs `aarch64`), downloads the latest stable Compose V2 binary, and sets it up locally in `~/.docker/cli-plugins/docker-compose`.
  - Added a force-remove container step (`docker rm -f`) in [.github/workflows/main.yml](file:///Users/rubenalves/Documents/repos/_school/iot/final/.github/workflows/main.yml) targeting the conflicting container ID `45d697b020602e7a04049c406bf768092fcea206f3861db4a00ad30d90612164` and any duplicate `ai_service` container names.
  - This solves both `'ContainerConfig'` key errors (from legacy v1 compatibility), `unknown shorthand flag: -f` errors (when the V2 plugin is missing from the VPS Docker installation), and naming conflicts from previous crashed deployments.

## 2026-05-30: Unified Domain and Subdomain Migration for Cloudflare

### Nginx Routing Configuration (`/deployments/nginx`)
- **[Unified Route Proxy]**: Modified [smartlock.conf](file:///Users/rubenalves/Documents/repos/_school/iot/final/deployments/nginx/smartlock.conf) to append a `location /api/` routing block forwarding requests to the Go backend (`localhost:8080`).
- **[Old Subdomain Deletion]**: Deleted [api.conf](file:///Users/rubenalves/Documents/repos/_school/iot/final/deployments/nginx/api.conf) as we consolidated the API onto `smartlock.raiiaa.dev/api`.

### ESP32 Firmware (`/arduino-ide`)
- **[Endpoints Migration]**: Modified [main.ino](file:///Users/rubenalves/Documents/repos/_school/iot/final/arduino-ide/main/main.ino) to replace all HTTP requests to the old `api.smartlock.raiiaa.dev` second-level subdomain with `smartlock.raiiaa.dev`.
- **[MQTT Server Route]**: Changed `MQTT_SERVER` variable in [main.ino](file:///Users/rubenalves/Documents/repos/_school/iot/final/arduino-ide/main/main.ino) to `mqtt.raiiaa.dev` (to be set up as a Cloudflare DNS-only record resolving directly to the DigitalOcean VPS to allow raw TCP connection on port 1883).

## 2026-05-30: Prettified AI Model Evaluation View with File Upload Support

### Go Backend (`/backend`)
- **[Direct CSV Evaluation]**: Modified [main.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/cmd/api/main.go) to import `"strings"` and update the `POST /api/ai/evaluate` endpoint. If the target dataset path does not exist on disk, it checks whether the string parameter contains CSV column headers (like `feature1` or `fails`). If it does, it forwards the content directly to the gRPC AI Service evaluator instead of falling back to default synthetic content.

### Vue Frontend (`/frontend`)
- **[Aesthetics & Layout Integration]**: Refactored [AiEvaluationView.vue](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/views/Pages/AiEvaluationView.vue) to wrap its contents in `<AdminLayout>`, utilize `<PageBreadcrumb>`, and group configuration controls and outputs into `<ComponentCard>` containers.
- **[Local File Upload Toggle]**: Implemented a toggle button switcher to select between **Ficheiro Local (Local CSV upload)** and **Caminho no Servidor (Server Path)**.
- **[Drag-and-Drop & FileReader]**: Built a custom drag-and-drop file selector in [AiEvaluationView.vue](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/views/Pages/AiEvaluationView.vue) that parses selected local `.csv` files via `FileReader.readAsText()` and submits their raw string contents directly to the backend.
- **[Evaluation Metrics & Heatmap]**: Expanded metric displays to separately showcase **Binary Metrics** (Anomaly threat detection) and **Macro Metrics** (multi-class severity levels) side-by-side. Designed a heatmapped confusion matrix visualization grid dynamically matching predicted and actual classes.
- **[Evaluation Composable Typings]**: Updated [useAiEvaluation.ts](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/composables/useAiEvaluation.ts) to define full metrics interface typing mapping the backend `EvaluationResult` model fields (`metrics`, `binary_metrics`, and `confusion_matrix`). Shifted the fetch route to a relative path `/api/ai/evaluate` to align with the Nginx reverse proxy configuration.
- **[History Table Formatting]**: Patched [EvaluationHistoryTable.vue](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/components/common/EvaluationHistoryTable.vue) to resolve self-import syntax errors. Styled the history logs with standard premium Tailwind classes matching the style of `UsersView.vue` and supporting dark-mode.
- **[Local Dev Proxy Target]**: Modified proxy settings in [vite.config.ts](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/vite.config.ts) from `http://main-server:8080` (resolvable only inside Docker) to `http://localhost:8080` to support standard local development on the host system without routing failures.
- **[Container Proxy Route]**: Configured `http-server` in [Dockerfile](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/Dockerfile) with the proxy flag `-P http://main-server:8080` to forward unresolved `/api` requests to the Go backend container inside the Docker network.

## 2026-05-31: Subdomain Routing Split and CORS Integration

### Nginx Routing Configuration (`/deployments/nginx`)
- **[Subdomains Split]**: Modified [smartlock.conf](file:///Users/rubenalves/Documents/repos/_school/iot/final/deployments/nginx/smartlock.conf) to split the single server block into three separate subdomains:
  - `smartlock.raiiaa.dev` -> Frontend (port 3000)
  - `smartlock-api.raiiaa.dev` -> Backend REST API (port 8080)
  - `smartlock-influx.raiiaa.dev` -> InfluxDB UI (port 8086)

### Go Backend (`/backend`)
- **[CORS Middleware]**: Registered and implemented a custom CORS middleware in [main.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/cmd/api/main.go) to allow cross-origin requests from the frontend subdomain (`smartlock.raiiaa.dev`) and handles OPTIONS preflight requests.

### Vue Frontend (`/frontend`)
- **[NEW] [Dynamic Endpoint Config]**: Created [config.ts](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/config.ts) to define a dynamic `API_BASE_URL` (resolving to `/api` in development, and defaulting to `https://smartlock-api.raiiaa.dev/api` in production).
- **[API Endpoint Migration]**: Updated [Device.vue](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/components/device/Device.vue) and [UsersView.vue](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/views/Tables/UsersView.vue) to query `API_BASE_URL` instead of hardcoded relative `/api/` paths.

### ESP32 Firmware (`/arduino-ide`)
- **[Endpoints Migration]**: Modified [main.ino](file:///Users/rubenalves/Documents/repos/_school/iot/final/arduino-ide/main/main.ino) to update the HTTP requests host from `smartlock.raiiaa.dev` to the new dedicated backend subdomain `smartlock-api.raiiaa.dev`.

## 2026-05-31: InfluxDB Token Alignment

### Deployments (`/deployments`)
- **[Token Mismatch Fix]**: Modified [docker-compose.yml](file:///Users/rubenalves/Documents/repos/_school/iot/final/deployments/docker-compose.yml) and [docker-compose.prod.yml](file:///Users/rubenalves/Documents/repos/_school/iot/final/deployments/docker-compose.prod.yml) to set the `DOCKER_INFLUXDB_INIT_ADMIN_TOKEN` variable to match the client `INFLUXDB_TOKEN` value (`"ZYPGu_Lu6NaP8M4iT5_TLx1xSZag9sAbR9i2vH8zr7P253VcxIMuXHbbkYagn2bOzfVZCRpsrIH5_77r3G1Mag=="`). This resolves 401 unauthorized errors on first-time initialization of database volumes.

### Go Backend (`/backend`)
- **[Configuration Default Update]**: Modified [config.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/config/config.go) to update the `envDefault` value for `InfluxDBToken` to align with the same token string (`"ZYPGu_Lu6NaP8M4iT5_TLx1xSZag9sAbR9i2vH8zr7P253VcxIMuXHbbkYagn2bOzfVZCRpsrIH5_77r3G1Mag=="`).

## 2026-05-31: System Architecture Documentation

### Docs (`/docs`)
- **[NEW] [AI Service Documentation]**: Created [IIA.md](file:///Users/rubenalves/Documents/repos/_school/iot/final/docs/IIA.md) to document the neural network architecture, data pre-processing, training procedures, and gRPC/RabbitMQ communication loops across Theoretical, Technical, and Pedagogical dimensions.
- **[NEW] [IoT Firmware Documentation]**: Created [IOT.md](file:///Users/rubenalves/Documents/repos/_school/iot/final/docs/IOT.md) to document sensors/actuators integration, local NVRAM caching strategies, physical polling optimization, and web setup configurations in both simulated PlatformIO and production Arduino IDE sketches.
- **[NEW] [Concurrency and Distribution Documentation]**: Created [PCD.md](file:///Users/rubenalves/Documents/repos/_school/iot/final/docs/PCD.md) to document the concurrent patterns implemented in the Go backend (Goroutines, channels, mutexes) and the end-to-end data flow trajectory across REST, gRPC, MQTT, and RabbitMQ protocols.

## 2026-05-31: Alignment of API Endpoints to Production Domain

### Vue Frontend (`/frontend`)
- **[Config Prefix Fix]**: Modified [config.ts](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/config.ts) to append the `/api` prefix to the production `API_BASE_URL` value (`https://smartlock-api.raiiaa.dev/api`).
- **[Composables Refactor]**: Refactored [useAiEvaluation.ts](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/composables/useAiEvaluation.ts) and [useAiRetrain.ts](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/composables/useAiRetrain.ts) to dynamically load the backend target via `API_BASE_URL` instead of relying on hardcoded URL string patterns.
- **[Door Control Fix]**: Modified [Device.vue](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/components/device/Device.vue) to update the HTTP POST target for remote door control actions from the invalid path `${API_BASE_URL}/device/door` to the correct backend route path `${API_BASE_URL}/door/unlock`.

### Simulated ESP32 Firmware (`/firmware`)
- **[Endpoints Migration]**: Modified [main.cpp](file:///Users/rubenalves/Documents/repos/_school/iot/final/firmware/src/main.cpp) to point its simulated HTTP client API destinations (`/api/users`, `/api/users/{uid}`, `/api/health`) and the remote MQTT server broker targets to the production subdomains `smartlock-api.raiiaa.dev` and `mqtt.raiiaa.dev`.

## 2026-05-31: Fixed Production Backend /api/health and Frontend SPA Fallback

### Go Backend (`/backend`)
- **[Health Route Registration]**: Modified [main.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/cmd/api/main.go) to restore both `r.Get("/api/health")` (returning real-time backing service connectivity status and latency from the monitor loop) and `r.Get("/api/metrics/health")` (querying time-series historical status pings from InfluxDB).

### Vue Frontend (`/frontend`)
- **[Production Server Upgrade]**: Modified the [Dockerfile](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/Dockerfile) to replace the `http-server` package with Vercel's standard `serve` package. Configured it to run `serve -s dist -l 8080` in the entrypoint command. The `-s` / `--single` flag ensures history API routing works natively in production, redirecting missing page refresh paths (such as `/ai/evaluation`) back to `index.html`.

## 2026-06-01: Fixed Device Control Reactivity, Connectivity Status, and State Binding

### Go Backend (`/backend`)
- **[Models Update]**: Modified [models.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/models/models.go) to add `Timestamp` (`time.Time`) and `IsOnline` (`bool`) fields to the `SensorPayload` struct.
- **[Repository Query Update]**: Modified [repository.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/domain/telemetry/repository.go) to select the `timestamp` column and scan it into `Timestamp` inside `GetLatest` db queries.
- **[Dynamic Connectivity Check]**: Refactored `GetLatestTelemetry` inside [service.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/domain/telemetry/service.go) to check if the timestamp of the latest telemetry is within a 15-second threshold, dynamically setting the `IsOnline` property of the returned JSON payload.

### Vue Frontend (`/frontend`)
- **[Rereactive Binds and Bads]**: Modified [Device.vue](file:///Users/rubenalves/Documents/repos/_school/iot/final/frontend/src/components/device/Device.vue):
  - Bind "Conetividade Hardware" card badge and styling classes to `deviceData?.is_online` status instead of the incorrect `status === 'online'` comparison.
  - Convert `isUnlocked` from a standard ref variable to a Vue `computed` property deriving from `deviceData` event states (`access_granted`, `status_change` with `UNLOCKED`) and manual override actions.
  - Disable the remote unlock button whenever the device goes offline (`!deviceData?.is_online`).

### ESP32 Firmware (`/arduino-ide` & `/firmware`)
- **[Device Telemetry Status Sync]**: Updated `sendTelemetry` in both [main.ino](file:///Users/rubenalves/Documents/repos/_school/iot/final/arduino-ide/main/main.ino) and [main.cpp](file:///Users/rubenalves/Documents/repos/_school/iot/final/firmware/src/main.cpp) to populate the `"status"` field with the actual current state of the lock (`isLocked ? "LOCKED" : "UNLOCKED"`) instead of writing empty strings, ensuring heartbeats always convey the current lock state.

## 2026-06-02: Fix Inline CSV Evaluation for Large Datasets

### Go Backend (`/backend`)
- **[Inline CSV Check Bug Fix]**: Modified [main.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/cmd/api/main.go) to check for inline CSV properties (presence of newlines or length > 255) prior to running `os.Stat`. This prevents `ENAMETOOLONG` errors when the frontend uploads large CSV payloads directly as strings to the `/api/ai/evaluate` endpoint.

