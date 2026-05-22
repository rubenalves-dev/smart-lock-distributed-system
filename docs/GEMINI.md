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
