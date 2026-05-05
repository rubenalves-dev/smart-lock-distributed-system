# Variables
PROTO_DIR = api/proto
GO_OUT_DIR = backend/internal/gen
PY_OUT_DIR = ai-service/src/gen

# Help command (default)
help:
	@echo "Usage:"
	@echo "  make proto      - Generate Go and Python code from .proto files"
	@echo "  make clean      - Remove generated files"

# Create directories and generate code
proto:
	@echo "Creating output directories..."
	mkdir -p $(GO_OUT_DIR)
	mkdir -p $(PY_OUT_DIR)

	@echo "Generating Go code..."
	protoc --proto_path=$(PROTO_DIR) \
		--go_out=$(GO_OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GO_OUT_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto

	@echo "Generating Python code..."
	python3 -m grpc_tools.protoc --proto_path=$(PROTO_DIR) \
		--python_out=$(PY_OUT_DIR) \
		--grpc_python_out=$(PY_OUT_DIR) \
		$(PROTO_DIR)/*.proto

	@echo "Done! Generated files are in $(GO_OUT_DIR) and $(PY_OUT_DIR)"

clean:
	rm -rf $(GO_OUT_DIR)/*
	rm -rf $(PY_OUT_DIR)/*


# Start the whole system
up:
	docker-compose -f ./deployments/docker-compose.yml -p smartlock up --build

# Stop the system
down:
	docker-compose down -f ./deployments/docker-compose.yml -p smartlock

# View logs
logs:
	docker-compose logs -f ./deployments/docker-compose.yml -p smartlock

# PlatformIO run
pio-run:
	cd ./firmware && pio run
