# 🛡️ Intelligent IoT Smart Lock

A complex distributed system for an intelligent security lock. This project integrates IoT sensors (ESP32), AI behavioral analysis (Python), and a concurrent microservice backend (Go).

## 🚀 Quick Start

1. **Prerequisites**: Ensure you have [Docker Desktop](https://www.docker.com/products/docker-desktop/) installed.
2. **Setup**: Run `make proto` to generate communication code.
3. **Launch**: Run `make up` to start the backend, AI, and Broker.

## 🛠️ Tooling Requirements (For Protoc)

To modify `.proto` files and regenerate code, team members must install:

- **Protobuf Compiler**: `protoc` (Download from GitHub releases)
- **Go Plugins**:
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  ```
- **Python Plugins**: `pip install grpcio-tools`

## 📂 Repository Structure

- `/api`: The "Source of Truth" (Protobuf definitions).
- `/backend`: Go microservices & Distributed logic.
- `/ai-service`: Python AI behavioral classification.
- `/firmware`: ESP32 C++ code & Local Web UI.
- `/deployments`: Docker Compose & Environment configs.
