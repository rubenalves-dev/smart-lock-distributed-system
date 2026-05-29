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
	// 1. If it is a heartbeat event, publish to RabbitMQ heartbeat queue and bypass synchronous DB/AI actions
	if payload.Event == "heartbeat" {
		payloadBytes, err := json.Marshal(payload)
		if err == nil && s.rabbitClient != nil {
			go func() {
				if err := s.rabbitClient.PublishHeartbeat(payloadBytes); err != nil {
					log.Printf("Failed to publish heartbeat event to RabbitMQ: %v\n", err)
				}
			}()
		}
		return nil
	}

	// 2. Intercept and handle RFID card validation dynamically
	if payload.RfidUID != "" {
		u, err := s.userService.GetUserByUID(ctx, payload.RfidUID)
		if err != nil {
			log.Printf("Failed to fetch user by RFID: %v\n", err)
		}

		if u == nil {
			// First time card is present
			log.Printf("First time scanning card %s. Storing in database and logging event.\n", payload.RfidUID)
			
			// Store the uid in the DB
			_, err = s.userService.RegisterTagIfNotExists(ctx, payload.RfidUID)
			if err != nil {
				log.Printf("Failed to automatically register RFID tag %s: %v\n", payload.RfidUID, err)
			}

			payload.Event = "new_card"
			payload.Details = "First scan of RFID card. Registered as pending."
			payload.Status = "Pending"
		} else {
			// Not the first time, check if accepted on database
			if u.IsAccepted && !u.IsBlocked {
				// Unlock the door
				if s.mqttClient != nil {
					go func() {
						if err := s.mqttClient.PublishOpenDoor(); err != nil {
							log.Printf("Failed to publish open door command to MQTT: %v\n", err)
						}
					}()
				}
				payload.Event = "access_granted"
				payload.Details = "Access granted: authorized card UID " + payload.RfidUID
				payload.Status = "Success"
			} else {
				// Do not unlock the door
				payload.Event = "access_denied"
				if u.IsBlocked {
					payload.Details = "Access denied: card blocked UID " + payload.RfidUID
					payload.Status = "Blocked"
				} else {
					payload.Details = "Access denied: card pending activation UID " + payload.RfidUID
					payload.Status = "Pending"
				}
			}
		}

		// Save telemetry log to Postgres
		if err := s.repo.Save(ctx, payload); err != nil {
			log.Printf("Failed to save telemetry to DB: %v\n", err)
		}
	} else {
		// Non-RFID event, save telemetry log to Postgres
		if err := s.repo.Save(ctx, payload); err != nil {
			log.Printf("Failed to save telemetry to DB: %v\n", err)
		}
	}

	// Marshal the final updated payload for RabbitMQ and AI prediction
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal updated sensor payload: %v\n", err)
		return err
	}

	// 4. Publish to RabbitMQ sensor events
	if s.rabbitClient != nil {
		go func() {
			if err := s.rabbitClient.PublishSensorEvent(payloadBytes); err != nil {
				log.Printf("Failed to publish sensor event to RabbitMQ: %v\n", err)
			}
		}()
	}

	// 5. Request AI prediction
	if s.aiService != nil {
		go func() {
			classification, confidence, recommendation, err := s.aiService.PredictSeverity(context.Background(), payload)
			if err != nil {
				log.Printf("AI Service PredictSeverity error: %v\n", err)
				return
			}
			log.Printf("AI Service Response: classification=%d, confidence=%.2f, recommendation=%s\n",
				classification, confidence, recommendation)
			
			// AI severity response door unlock - only if it is actually granted
			if payload.Event == "access_granted" && s.mqttClient != nil {
				go func() {
					if err := s.mqttClient.PublishOpenDoor(); err != nil {
						log.Printf("Failed to publish open door command to MQTT: %v\n", err)
					}
				}()
			}

			if classification >= 2 && s.rabbitClient != nil { // Severity 2 is Suspicious, 3 is High
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

