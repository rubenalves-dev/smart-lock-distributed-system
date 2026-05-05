# 📑 API Contracts

This directory contains the definitions for all service-to-service communication.

### 📜 service.proto

This file defines:

- **Data Models**: `SensorEvent`, `SecurityAssessment`.
- **gRPC Services**: `AIService` and `LockService`.

### 🔄 Generation

**Do not edit the generated files in `/gen` folders manually.**
Always modify the `.proto` file here and run `make proto` from the root. This ensures that the Go server and Python AI stay synchronized.
