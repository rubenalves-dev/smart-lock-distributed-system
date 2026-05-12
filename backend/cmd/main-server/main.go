package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq" // Driver vital para o Postgres
	"github.com/rubenalves-dev/smart-lock-distributed-system/broker"
	smartlock "github.com/rubenalves-dev/smart-lock-distributed-system/internal/gen"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// 1. Canal para receber dados do MQTT
	telemetryChan := make(chan models.SensorPayload, 100)

	// 2. Iniciar o Subscriber MQTT
	go broker.StartSubscriber("mosquitto_broker", telemetryChan)

	// 3. Configurar ligação ao Postgres
	// Usamos o nome do serviço definido no docker-compose: postgres_db
	dbURL := "host=postgres_db user=user password=password dbname=lock_db sslmode=disable"
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("Erro ao configurar base de dados: %v\n", err)
	}
	defer db.Close()

	// 4. Configurar ligação gRPC para a IA
	conn, err := grpc.NewClient("ai-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("Failed to connect to AI service: %v\n", err)
	}
	aiClient := smartlock.NewAIServiceClient(conn)

	// 5. Configurar RabbitMQ
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		// Fallback para o nome do serviço no Docker
		rabbitURL = "amqp://guest:guest@rabbitmq:5672/"
	}
	rabbit, err := broker.NewRabbitMQ(rabbitURL)
	if err != nil {
		fmt.Printf("Failed to connect to RabbitMQ: %v\n", err)
	}

	// 6. LOOP PRINCIPAL - Onde a magia acontece
	go func() {
		for event := range telemetryChan {
			fmt.Printf("📥 Processing %s from %s (Value: %v)\n", event.EventType, event.DeviceID, event.Value)

			// A. GRAVAR NO POSTGRES
			query := `INSERT INTO sensor_data (device_id, event_type, value, created_at) VALUES ($1, $2, $3, $4)`
			_, err := db.Exec(query, event.DeviceID, event.EventType, event.Value, time.Now())
			if err != nil {
				fmt.Printf("❌ Erro Postgres: %v\n", err)
			} else {
				fmt.Println("✅ Guardado no Postgres com sucesso!")
			}

			// B. ENVIAR PARA O RABBITMQ (Para a IA processar em background)
			bytes, _ := json.Marshal(event)
			if rabbit != nil {
				go rabbit.PublishSensorEvent(bytes)
			}

			// C. CHAMADA gRPC (Análise em tempo real)
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
	}()

	// 7. Servidor de Health Check e Graceful Shutdown
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Server is healthy"))
	})

	go http.ListenAndServe(":8080", mux)
	fmt.Println("🚀 Main Server is running on port 8080")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")
}
