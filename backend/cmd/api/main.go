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

type SyntheticData struct {
	Feature1 float64 `json:"feature1"`
	Feature2 float64 `json:"feature2"`
}

type EvaluateRequest struct {
	DatasetPath   string          `json:"dataset_path"`
	SyntheticData []SyntheticData `json:"synthetic_data"`
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	telemetryChan := make(chan models.SensorPayload, 100)

	dbClient, err := core.NewPostgresClient(cfg.PostgresURL)
	if err != nil {
		log.Fatalf("Critical: Failed to connect to PostgreSQL: %v", err)
	}
	defer dbClient.Close()

	influxClient, err := core.NewInfluxClient(cfg.InfluxDBURL, cfg.InfluxDBToken)
	if err != nil {
		log.Fatalf("Critical: Failed to connect to InfluxDB: %v", err)
	}
	defer influxClient.Close()

	rabbitClient, err := core.NewRabbitMQClient(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Critical: Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitClient.Close()

	var grpcConn *grpc.ClientConn
	for i := range 15 {
		grpcConn, err = grpc.Dial(cfg.AIServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			break
		}
		log.Printf("... retrying... (%d/15): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Critical: Failed to connect to AI Service: %v", err)
	}
	defer grpcConn.Close()

	mqttClient, err := core.NewMQTTClient(cfg.MQTTBroker, telemetryChan)
	if err != nil {
		log.Fatalf("Critical: Failed to connect to MQTT broker: %v", err)
	}
	defer mqttClient.Close()

	rawAIClient := smartlock.NewAIServiceClient(grpcConn)
	aiService := ai.NewGRPCClient(rawAIClient)

	userRepo := user.NewRepository(dbClient.DB)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	telemetryRepo := telemetry.NewRepository(dbClient.DB)
	telemetryService := telemetry.NewService(telemetryRepo, userService, rabbitClient, mqttClient, aiService)
	telemetryHandler := telemetry.NewHandler(telemetryService)

	healthMonitor := monitor.NewMonitor(dbClient, mqttClient, influxClient, rabbitClient, cfg.RabbitMQURL, cfg.InfluxDBOrg, cfg.InfluxDBBucket)
	monitorCtx, cancelMonitor := context.WithCancel(context.Background())
	defer cancelMonitor()
	go healthMonitor.Start(monitorCtx)

	go func() {
		for event := range telemetryChan {
			_ = telemetryService.Ingest(context.Background(), event)
		}
	}()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`{"status":"OK"}`)) })

	r.Post("/api/ai/evaluate", func(w http.ResponseWriter, r *http.Request) {
		var req EvaluateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Erro ao decodificar pedido", http.StatusBadRequest)
			return
		}

		pathToEvaluate := req.DatasetPath
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		// 1. Fallback se não houver caminho
		if pathToEvaluate == "" {
			pathToEvaluate = "/test_data.csv"
		}

		// 2. Se o ficheiro não existir, gera dados sintéticos na memória
		if _, err := os.Stat(pathToEvaluate); os.IsNotExist(err) {
			log.Println("Ficheiro não encontrado, a gerar dados sintéticos na memória...")
			// Cria um buffer em memória para simular o ficheiro
			content := "feature1,feature2\n0.5,0.2"

			// Enviamos a string 'content' diretamente para o gRPC
			result, err := aiService.EvaluateModel(ctx, content)
			if err != nil {
				log.Printf("Erro na IA: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(result)
			return
		}

		// 3. Caso o ficheiro exista, lê o conteúdo e envia
		contentBytes, err := os.ReadFile(pathToEvaluate)
		if err != nil {
			http.Error(w, "Erro ao ler ficheiro", http.StatusInternalServerError)
			return
		}

		// IMPORTANTE: Envia o conteúdo convertido para string
		result, err := aiService.EvaluateModel(ctx, string(contentBytes))
		if err != nil {
			log.Printf("Erro na IA: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	userHandler.RegisterRoutes(r)
	telemetryHandler.RegisterRoutes(r)

	srv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.Port), Handler: r}
	go srv.ListenAndServe()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	srv.Shutdown(context.Background())
}
