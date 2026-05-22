# Gemini Codebase Structural Changes Log

This file tracks all structural changes made by the AI Agent to keep the project documented and organized.

## Date: 2026-05-22

### Python AI Service (`/ai-service`)
- **[Dependency Update]**: Added `pika>=1.3.0` to [requirements.txt](file:///Users/rubenalves/Documents/repos/_school/iot/final/ai-service/requirements.txt) to support RabbitMQ messaging queue ingestion.
- **[Model Helpers]**: Added dataset generation, normalization, loading, and training helper functions to [model.py](file:///Users/rubenalves/Documents/repos/_school/iot/final/ai-service/src/model.py).
- **[Refactoring & Features]**: Refactored [main.py](file:///Users/rubenalves/Documents/repos/_school/iot/final/ai-service/src/main.py):
  - Created a baseline model on startup if not present to prevent service crashes.
  - Implemented the gRPC `PredictSeverity` endpoint to parse features (`fails`, `distance_cm`, `is_denied`) from real `SensorEvent` protobuf structures.
  - Implemented the gRPC `RetrainModel` endpoint to retrain the neural network on arbitrary dataset paths (CSV) and reload it dynamically in-memory with thread safety.
  - Implemented an asynchronous background daemon thread (`RabbitMQConsumer`) consuming events from the `sensor_events` queue and recording them into `data/sensor_events.csv` for self-supervised data compilation.

### Go Backend (`/backend`)
- **[Dependency Update]**: Added `github.com/go-chi/chi/v5` for REST routing and `github.com/caarlos0/env/v11` for configuration parsing to `go.mod`.
- **[Config Module]**: Created the [config.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/config/config.go) module using `caarlos0/env` to parse `RABBITMQ_URL`, `AI_SERVICE_ADDR`, and `PORT` variables from environment settings.
- **[Entrypoint Restructuring]**: Created the new API entry point at [main.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/cmd/api/main.go) and removed the old entry point `backend/cmd/main-server/main.go`.
- **[HTTP Routing & APIs]**: Implemented Chi router integration in `cmd/api/main.go` with `/health` and `/api/ai/retrain` endpoints. Fixes a bug where telemetry consumption blocked the HTTP server startup.
- **[Dockerfile Update]**: Updated the [Dockerfile](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/Dockerfile) to build the binary using the new entrypoint path `cmd/api/main.go`.
