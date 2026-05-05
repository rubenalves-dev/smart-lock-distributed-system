# 🐹 Go Backend (Main Server)

The brain of the distributed system. It orchestrates auth requests and bridges the Hardware with the AI.

### 🏗️ Architecture

- **gRPC Server**: Listens for internal calls from other services.
- **RabbitMQ Producer**: Sends sensor data to the AI asynchronously.
- **REST API**: Serves data to the Frontend/App.

### 💡 Must-Knows

- **Concurrency**: Use `goroutines` for the RabbitMQ consumer and `channels` to pipe data to the database.
- **Retries**: The server includes a retry loop for the DB and Broker connections (crucial for Docker startup).
