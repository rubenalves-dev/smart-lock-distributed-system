package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/config"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/core"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/domain/ai"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/domain/telemetry"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/domain/user"
	smartlock "github.com/rubenalves-dev/smart-lock-distributed-system/internal/gen"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/models"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/monitor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting backend with Config: Port=%d, AIAddr=%s, RabbitURL=%s\n", cfg.Port, cfg.AIServiceAddr, cfg.RabbitMQURL)

	// Create channel for MQTT telemetry events
	telemetryChan := make(chan models.SensorPayload, 100)

	// Initialize PostgreSQL Client
	dbClient, err := core.NewPostgresClient(cfg.PostgresURL)
	if err != nil {
		log.Fatalf("Critical: Failed to connect to PostgreSQL: %v", err)
	}
	defer dbClient.Close()

	// Initialize InfluxDB Client
	influxClient, err := core.NewInfluxClient(cfg.InfluxDBURL, cfg.InfluxDBToken)
	if err != nil {
		log.Fatalf("Critical: Failed to connect to InfluxDB: %v", err)
	}
	defer influxClient.Close()

	// Initialize RabbitMQ Client
	rabbitClient, err := core.NewRabbitMQClient(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Critical: Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitClient.Close()

	// Initialize gRPC connection to AI Service
	var grpcConn *grpc.ClientConn
	for i := range 15 {
		grpcConn, err = grpc.NewClient(cfg.AIServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			log.Println("Connected to AI service gRPC server successfully")
			break
		}
		log.Printf("Failed to connect to AI service gRPC, retrying in 2 seconds... (%d/15): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Critical: Failed to connect to AI Service: %v", err)
	}
	defer grpcConn.Close()

	// Initialize MQTT Mosquitto Client
	mqttClient, err := core.NewMQTTClient(cfg.MQTTBroker, telemetryChan)
	if err != nil {
		log.Fatalf("Critical: Failed to connect to MQTT broker: %v", err)
	}
	defer mqttClient.Close()

	// Initialize domains
	// AI Domain
	rawAIClient := smartlock.NewAIServiceClient(grpcConn)
	aiService := ai.NewGRPCClient(rawAIClient)

	// User Domain
	userRepo := user.NewRepository(dbClient.DB)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	// Telemetry Domain
	telemetryRepo := telemetry.NewRepository(dbClient.DB)
	telemetryService := telemetry.NewService(telemetryRepo, userService, rabbitClient, mqttClient, aiService)
	telemetryHandler := telemetry.NewHandler(telemetryService)

	// 7. Start Background Health Monitor
	healthMonitor := monitor.NewMonitor(
		dbClient,
		mqttClient,
		influxClient,
		rabbitClient,
		cfg.RabbitMQURL,
		cfg.InfluxDBOrg,
		cfg.InfluxDBBucket,
	)

	monitorCtx, cancelMonitor := context.WithCancel(context.Background())
	defer cancelMonitor()
	go healthMonitor.Start(monitorCtx)

	// 8. Start MQTT Event Consumer loop asynchronously
	go func() {
		for event := range telemetryChan {
			log.Printf("Processing MQTT event: %s from device: %s\n", event.Event, event.DeviceID)
			if err := telemetryService.Ingest(context.Background(), event); err != nil {
				log.Printf("Error ingesting telemetry event: %v\n", err)
			}
		}
	}()

	// 9. Start RabbitMQ Heartbeat Consumer loop asynchronously
	go func() {
		log.Println("Starting background RabbitMQ heartbeat consumer...")
		err := rabbitClient.ConsumeHeartbeats(context.Background(), func(body []byte) error {
			var payload models.SensorPayload
			if err := json.Unmarshal(body, &payload); err != nil {
				return err
			}
			log.Printf("Background saving heartbeat event from device: %s (uptime: %.1f)\n", payload.DeviceID, payload.Uptime)
			return telemetryRepo.Save(context.Background(), payload)
		})
		if err != nil {
			log.Printf("Failed to start RabbitMQ heartbeat consumer: %v\n", err)
		}
	}()

	// 9. Setup go-chi router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"OK"}`))
	})

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		statuses := healthMonitor.GetLatestStatuses()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(statuses)
	})

	// Register domain routes
	userHandler.RegisterRoutes(r)
	telemetryHandler.RegisterRoutes(r)

	// AI retrain endpoint
	r.Post("/api/ai/retrain", func(w http.ResponseWriter, r *http.Request) {
		type RetrainRequest struct {
			Epochs      int32  `json:"epochs"`
			DatasetPath string `json:"dataset_path"`
		}

		var req RetrainRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req.Epochs <= 0 {
			req.Epochs = 10
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		success, message, err := aiService.RetrainModel(ctx, req.Epochs, req.DatasetPath)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("AI error: %v", err),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": success,
			"message": message,
		})
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: r,
	}

	go func() {
		log.Printf("REST API server listening on :%d", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen and serve error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced shutdown: %v", err)
	}

	log.Println("Server stopped.")
	os.Exit(0)
}
