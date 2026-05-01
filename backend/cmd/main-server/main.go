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
	telemetryChan := make(chan models.SensorPayload, 100)

	go broker.StartSubscriber("192.168.1.219", telemetryChan)

	conn, err := grpc.NewClient("ai-service:50051", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("Failed to connect to AI service: %v\n", err)
	}
	aiClient := smartlock.NewAIServiceClient(conn)

	rabbit, err := broker.NewRabbitMQ("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		fmt.Printf("Failed to connect to RabbitMQ: %v\n", err)
	}

	for event := range telemetryChan {
		fmt.Printf("Processing %s from %s\n", event.EventType, event.DeviceID)
		bytes, err := json.Marshal(event)
		if err != nil {
			fmt.Printf("Failed to marshal event: %v\n", err)
			continue
		}

		go rabbit.PublishSensorEvent(bytes)

		if event.EventType == "vibration" {
			resp, err := aiClient.PredictSeverity(context.Background(), &smartlock.PredictSeverityRequest{
				Events: []*smartlock.SensorEvent{{
					VibrationIntensity: float32(event.Value),
					EntryMethod:        "vibration",
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
