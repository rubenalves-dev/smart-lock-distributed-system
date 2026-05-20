package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rubenalves-dev/smart-lock-distributed-system/broker"
	smartlock "github.com/rubenalves-dev/smart-lock-distributed-system/internal/gen"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/models"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	telemetryChan := make(chan models.SensorPayload, 100)
	go broker.StartSubscriber("mosquitto_broker", telemetryChan)

	conn, err := grpc.NewClient("ai-service:50051", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("Failed to connect to AI service: %v\n", err)
	}
	aiClient := smartlock.NewAIServiceClient(conn)

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://user:password@localhost:5672/"
	}
	rabbit, err := broker.NewRabbitMQ(rabbitURL)
	if err != nil {
		fmt.Printf("Failed to connect to RabbitMQ: %v\n", err)
	}

	for event := range telemetryChan {
		fmt.Printf("Processing %s from %s\n", event.Event, event.DeviceID)
		bytes, err := json.Marshal(event)
		if err != nil {
			fmt.Printf("Failed to marshal event: %v\n", err)
			continue
		}

		go rabbit.PublishSensorEvent(bytes)

		if event.Event != "heartbeat" {
			resp, err := aiClient.PredictSeverity(ctx, &smartlock.PredictSeverityRequest{
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

			if err == nil && resp.Classification >= smartlock.Severity_SEVERITY_SUSPICIOUS {
				fmt.Printf("⚠️ ALERT: AI classified this as %v\n", resp.Classification)
			}
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Server is healthy"))
	})

	quit := make(chan os.Signal, 1)
	go http.ListenAndServe(":8080", mux)

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	// Perform any necessary cleanup before shutting down
	// This could include closing database connections, stopping background tasks, etc.

	fmt.Println("Shutting down server...")
	fmt.Println("Server stopped.")

	os.Exit(0)
}
