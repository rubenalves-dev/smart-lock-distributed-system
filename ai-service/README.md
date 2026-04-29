# 🐍 AI Analysis Service (Python)

Microservice dedicated to classifying behavior based on lock interactions.

### 🧠 Features

- **Classification**: Labels data as `OK`, `Irregular`, `Suspicious`, or `Dangerous`.
- **gRPC Endpoint**: `PredictSeverity` for real-time analysis.
- **Async Processing**: Consumes from the RabbitMQ `sensor_events` queue.

### 🛠️ Setup

- Edit `requirements.txt` to add ML libraries (Scikit-learn/TensorFlow).
- Use `gen/` path for gRPC imports.
