package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rubenalves-dev/smart-lock-distributed-system/broker"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/config"
	smartlock "github.com/rubenalves-dev/smart-lock-distributed-system/internal/gen"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Starting backend with Config: Port=%s, AIAddr=%s, RabbitURL=%s\n", cfg.Port, cfg.AIServiceAddr, cfg.RabbitMQURL)

	// gRPC connection to AI Service using modern grpc.NewClient
	conn, err := grpc.NewClient(cfg.AIServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("Failed to connect to AI service: %v\n", err)
	}
	defer conn.Close()
	aiClient := smartlock.NewAIServiceClient(conn)

	// RabbitMQ initialization with retry loop
	var rabbit *broker.RabbitMQClient
	for i := 1; i <= 10; i++ {
		rabbit, err = broker.NewRabbitMQ(cfg.RabbitMQURL)
		if err == nil {
			fmt.Println("Successfully connected to RabbitMQ broker")
			break
		}
		fmt.Printf("Failed to connect to RabbitMQ (attempt %d/10): %v. Retrying in 3 seconds...\n", i, err)
		time.Sleep(3 * time.Second)
	}

	// Start Telemetry subscriber
	telemetryChan := make(chan models.SensorPayload, 100)
	go broker.StartSubscriber("mosquitto_broker", telemetryChan)

	// Run Telemetry processing in a separate goroutine (non-blocking)
	go func() {
		for event := range telemetryChan {
			fmt.Printf("Processing %s from %s\n", event.Event, event.DeviceID)
			bytes, err := json.Marshal(event)
			if err != nil {
				fmt.Printf("Failed to marshal event: %v\n", err)
				continue
			}

			if rabbit != nil {
				go rabbit.PublishSensorEvent(bytes)
			}

			if event.Event != "heartbeat" && aiClient != nil {
				resp, err := aiClient.PredictSeverity(context.Background(), &smartlock.PredictSeverityRequest{
					Events: []*smartlock.SensorEvent{{
						DeviceId:   event.DeviceID,
						Event:      event.Event,
						Detail:     event.Details,
						Status:     event.Status,
						DistanceCm: event.DistanceCm,
						LightLevel: int32(event.LightLevel),
						Fails:      int32(event.Fails),
						User:       event.User,
						Rssi:       int32(event.RSSI),
						Uptime:     event.Uptime,
					}},
				})

				if err == nil {
					fmt.Printf("AI response: classification=%v, confidence=%.2f, recommendation=%s\n",
						resp.Classification, resp.Confidence, resp.Recommendation)
					if resp.Classification >= smartlock.Severity_SEVERITY_SUSPICIOUS {
						fmt.Printf("⚠️ ALERT: AI classified this as %v\n", resp.Classification)
					}
				} else {
					fmt.Printf("Failed to predict severity: %v\n", err)
				}
			}
		}
	}()

	// chi Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"OK"}`))
	})

	r.Post("/api/ai/retrain", func(w http.ResponseWriter, r *http.Request) {
		type RetrainRequest struct {
			Epochs      int32  `json:"epochs"`
			DatasetPath string `json:"dataset_path"`
		}

		var req RetrainRequest
		// Attempt to decode, if fails or body is empty, we fall back to defaults
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req.Epochs <= 0 {
			req.Epochs = 10
		}

		if aiClient == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"success":false,"message":"AI client is not connected"}`))
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := aiClient.RetrainModel(ctx, &smartlock.RetrainModelRequest{
			Epochs:      req.Epochs,
			DatasetPath: req.DatasetPath,
		})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("gRPC error: %v", err),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	})

	// Server shutdown setup
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Printf("HTTP Server is listening on port %s...\n", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server listen error: %v\n", err)
		}
	}()

	<-quit
	fmt.Println("Shutting down HTTP server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Server shutdown error: %v\n", err)
	}
	fmt.Println("Server stopped.")
}
