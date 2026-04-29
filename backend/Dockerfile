# Stage 1: Build
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main-server ./cmd/main-server/main.go

# Stage 2: Run
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main-server .
# Expose REST API and gRPC ports
EXPOSE 8080 50052
CMD ["./main-server"]
