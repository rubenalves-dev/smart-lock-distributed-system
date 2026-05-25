package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/core"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/domain/ai"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/domain/user"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/models"
)


type Service struct {
	repo         Repository
	userService  *user.Service
	rabbitClient *core.RabbitMQClient
	mqttClient   *core.MQTTClient
	aiService    ai.AIService
}

func NewService(
	repo Repository,
	userService *user.Service,
	rabbitClient *core.RabbitMQClient,
	mqttClient *core.MQTTClient,
	aiService ai.AIService,
) *Service {
	return &Service{
		repo:         repo,
		userService:  userService,
		rabbitClient: rabbitClient,
		mqttClient:   mqttClient,
		aiService:    aiService,
	}
}

func (s *Service) Ingest(ctx context.Context, payload models.SensorPayload) error {
	// Marshal the sensor payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal sensor payload: %v\n", err)
		return err
	}

	// 1. If it is a heartbeat event, publish to RabbitMQ heartbeat queue and bypass synchronous DB/AI actions
	if payload.Event == "heartbeat" {
		if s.rabbitClient != nil {
			go func() {
				if err := s.rabbitClient.PublishHeartbeat(payloadBytes); err != nil {
					log.Printf("Failed to publish heartbeat event to RabbitMQ: %v\n", err)
				}
			}()
		}
		return nil
	}

	// 2. Save telemetry log to Postgres synchronously for real events
	if err := s.repo.Save(ctx, payload); err != nil {
		log.Printf("Failed to save telemetry to DB: %v\n", err)
	}

	// 3. Automatically register new RFID tag if it's access_granted
	if payload.Event == "access_granted" && payload.RfidUID != "" {
		_, err := s.userService.RegisterTagIfNotExists(ctx, payload.RfidUID)
		if err != nil {
			log.Printf("Failed to automatically register RFID tag %s: %v\n", payload.RfidUID, err)
		} else {
			log.Printf("RFID tag %s checked/registered successfully\n", payload.RfidUID)
		}
	}

	// 4. Publish to RabbitMQ sensor events
	if s.rabbitClient != nil {
		go func() {
			if err := s.rabbitClient.PublishSensorEvent(payloadBytes); err != nil {
				log.Printf("Failed to publish sensor event to RabbitMQ: %v\n", err)
			}
		}()
	}

	// 5. Request AI prediction if not heartbeat
	if s.aiService != nil {
		go func() {
			classification, confidence, recommendation, err := s.aiService.PredictSeverity(context.Background(), payload)
			if err != nil {
				log.Printf("AI Service PredictSeverity error: %v\n", err)
				return
			}
			log.Printf("AI Service Response: classification=%d, confidence=%.2f, recommendation=%s\n",
				classification, confidence, recommendation)
			go func() {
				if err := s.mqttClient.PublishOpenDoor(); err != nil {
					log.Printf("Failed to publish open door command to MQTT: %v\n", err)
				}
			}()

			if classification >= 2 { // Severity 2 is Suspicious, 3 is High
				go func() {
					if err := s.rabbitClient.PublishRequestMFA(); err != nil {
						log.Printf("Failed to publish MFA request to RabbitMQ: %v\n", err)
					}
				}()
			}
		}()
	}

	return nil
}

func (s *Service) GetLatestTelemetry(ctx context.Context, deviceID string) (*models.SensorPayload, error) {
	return s.repo.GetLatest(ctx, deviceID)
}

func (s *Service) UnlockDoor(ctx context.Context) error {
	if s.mqttClient == nil {
		return fmt.Errorf("MQTT client is nil")
	}
	return s.mqttClient.PublishOpenDoor()
}

