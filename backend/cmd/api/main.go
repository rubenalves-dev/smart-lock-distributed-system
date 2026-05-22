package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	_ "github.com/lib/pq"
	"github.com/rubenalves-dev/smart-lock-distributed-system/broker"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/config"
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

	telemetryChan := make(chan models.SensorPayload, 100)

	// Connect to PostgreSQL with retry
	var db *sql.DB
	for {
		db, err = sql.Open("postgres", cfg.PostgresURL)
		if err == nil {
			err = db.Ping()
			if err == nil {
				log.Println("Connected to PostgreSQL DB successfully")
				break
			}
		}
		log.Printf("Failed to connect to PostgreSQL (URL: %s), retrying in 2 seconds...: %v", cfg.PostgresURL, err)
		time.Sleep(2 * time.Second)
	}
	defer db.Close()

	// Connect to MQTT Mosquitto with retry
	var mqttClient mqtt.Client
	for {
		mqttClient, err = broker.StartSubscriber(cfg.MQTTBroker, telemetryChan)
		if err == nil {
			log.Println("Connected to MQTT broker successfully")
			break
		}
		log.Printf("Failed to connect to MQTT broker (%s), retrying in 2 seconds...: %v", cfg.MQTTBroker, err)
		time.Sleep(2 * time.Second)
	}

	// Connect to InfluxDB with retry
	influxClient := influxdb2.NewClient(cfg.InfluxDBURL, cfg.InfluxDBToken)
	defer influxClient.Close()
	for {
		ok, err := influxClient.Ping(context.Background())
		if err == nil && ok {
			log.Println("Connected to InfluxDB successfully")
			break
		}
		log.Printf("Failed to connect to InfluxDB (URL: %s), retrying in 2 seconds...: %v", cfg.InfluxDBURL, err)
		time.Sleep(2 * time.Second)
	}

	// Connect to RabbitMQ with retry
	var rabbit *broker.RabbitMQClient
	for {
		rabbit, err = broker.NewRabbitMQ(cfg.RabbitMQURL)
		if err == nil {
			log.Println("Connected to RabbitMQ successfully")
			break
		}
		log.Printf("Failed to connect to RabbitMQ (URL: %s), retrying in 2 seconds...: %v", cfg.RabbitMQURL, err)
		time.Sleep(2 * time.Second)
	}
	defer rabbit.Close()

	// gRPC connection to AI Service with retry
	var grpcConn *grpc.ClientConn
	for {
		grpcConn, err = grpc.NewClient(cfg.AIServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			log.Println("Connected to AI service gRPC server successfully")
			break
		}
		log.Printf("Failed to connect to AI service gRPC, retrying in 2 seconds...: %v", err)
		time.Sleep(2 * time.Second)
	}
	defer grpcConn.Close()
	aiClient := smartlock.NewAIServiceClient(grpcConn)

	// Start Background Health Monitor
	healthMonitor := monitor.NewMonitor(
		db,
		mqttClient,
		influxClient,
		rabbit,
		cfg.RabbitMQURL,
		cfg.InfluxDBOrg,
		cfg.InfluxDBBucket,
	)

	monitorCtx, cancelMonitor := context.WithCancel(context.Background())
	defer cancelMonitor()
	go healthMonitor.Start(monitorCtx)

	// Start MQTT Event Consumer loop asynchronously
	go func() {
		for event := range telemetryChan {
			log.Printf("Processing %s from %s\n", event.Event, event.DeviceID)
			bytes, err := json.Marshal(event)
			if err != nil {
				log.Printf("Failed to marshal event: %v\n", err)
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
					log.Printf("AI response: classification=%v, confidence=%.2f, recommendation=%s\n",
						resp.Classification, resp.Confidence, resp.Recommendation)
					if resp.Classification >= smartlock.Severity_SEVERITY_SUSPICIOUS {
						log.Printf("⚠️ ALERT: AI classified this as %v\n", resp.Classification)
					}
				} else {
					log.Printf("Failed to predict severity: %v\n", err)
				}
			}
		}
	}()

	// Setup go-chi router
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
