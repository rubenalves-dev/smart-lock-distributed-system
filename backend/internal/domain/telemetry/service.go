package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/core"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/domain/ai"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/domain/mfa"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/domain/user"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/models"
)


type Service struct {
	repo         Repository
	userService  *user.Service
	rabbitClient *core.RabbitMQClient
	mqttClient   *core.MQTTClient
	aiService    ai.AIService
	mfaService   *mfa.Service
}

func NewService(
	repo Repository,
	userService *user.Service,
	rabbitClient *core.RabbitMQClient,
	mqttClient *core.MQTTClient,
	aiService ai.AIService,
	mfaService *mfa.Service,
) *Service {
	return &Service{
		repo:         repo,
		userService:  userService,
		rabbitClient: rabbitClient,
		mqttClient:   mqttClient,
		aiService:    aiService,
		mfaService:   mfaService,
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

			if err := s.repo.Save(ctx, payload); err != nil {
				log.Printf("Failed to save telemetry to DB: %v\n", err)
			}
		} else if payload.Event == "access_request" || payload.Event == "access_denied" {
			// Online access evaluation request
			if u.IsAccepted && !u.IsBlocked {
				var classification int32 = 0
				var confidence float32 = 1.0
				var recommendation = "Status normal. Access allowed."

				// Request AI prediction synchronously
				if s.aiService != nil {
					c, conf, rec, err := s.aiService.PredictSeverity(ctx, payload)
					if err != nil {
						log.Printf("AI Service PredictSeverity error: %v\n", err)
					} else {
						classification = c
						confidence = conf
						recommendation = rec
					}
				}

				if classification < 2 {
					// Unlock the door immediately
					if s.mqttClient != nil {
						go func() {
							if err := s.mqttClient.PublishOpenDoor(); err != nil {
								log.Printf("Failed to publish open door command to MQTT: %v\n", err)
							}
						}()
					}
					payload.Event = "access_granted"
					payload.Details = fmt.Sprintf("Access granted (AI evaluated OK): %s", recommendation)
					payload.Status = "Success"
				} else {
					// Hold door locked, trigger MFA request
					if s.mfaService != nil {
						_, err := s.mfaService.CreateRequest(ctx, payload.RfidUID, payload.DeviceID, payload.Fails, payload.DistanceCm, payload.LightLevel, int(classification), confidence, recommendation)
						if err != nil {
							log.Printf("Failed to create MFA request: %v\n", err)
						}
					}
					payload.Event = "mfa_pending"
					payload.Details = fmt.Sprintf("MFA verification required: %s", recommendation)
					payload.Status = "Pending"
				}
			} else {
				// Blocked or pending activation
				payload.Event = "access_denied"
				if u.IsBlocked {
					payload.Details = "Access denied: card blocked UID " + payload.RfidUID
					payload.Status = "Blocked"
				} else {
					payload.Details = "Access denied: card pending activation UID " + payload.RfidUID
					payload.Status = "Pending"
				}
			}

			if err := s.repo.Save(ctx, payload); err != nil {
				log.Printf("Failed to save telemetry to DB: %v\n", err)
			}
		} else {
			// Local access_granted/access_denied (e.g. offline cache) or registration events, save directly
			if err := s.repo.Save(ctx, payload); err != nil {
				log.Printf("Failed to save telemetry to DB: %v\n", err)
			}
		}
	} else {
		// Non-RFID event, save telemetry log to Postgres
		if err := s.repo.Save(ctx, payload); err != nil {
			log.Printf("Failed to save telemetry to DB: %v\n", err)
		}
	}

	// Ingest payload to RabbitMQ sensor events queue for asynchronously updating retraining logs
	payloadBytes, err := json.Marshal(payload)
	if err == nil && s.rabbitClient != nil {
		go func() {
			if err := s.rabbitClient.PublishSensorEvent(payloadBytes); err != nil {
				log.Printf("Failed to publish sensor event to RabbitMQ: %v\n", err)
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

