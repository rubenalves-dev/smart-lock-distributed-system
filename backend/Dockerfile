# Stage 1: Build
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o api-server ./cmd/api/main.go

# Stage 2: Run
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/api-server .
# Expose REST API and gRPC ports
EXPOSE 8080 50052
CMD ["./api-server"]
