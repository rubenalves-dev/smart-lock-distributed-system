# Gemini Codebase Structural Changes Log

This document tracks structural changes made to the repository code.

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

