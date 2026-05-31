package mfa

import (
	"context"
	"fmt"
	"log"

	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/core"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/domain/user"
)

type Service struct {
	repo        Repository
	userService *user.Service
	mqttClient  *core.MQTTClient
	wsHub       *core.Hub
}

func NewService(repo Repository, userService *user.Service, mqttClient *core.MQTTClient, wsHub *core.Hub) *Service {
	return &Service{
		repo:        repo,
		userService: userService,
		mqttClient:  mqttClient,
		wsHub:       wsHub,
	}
}

func (s *Service) CreateRequest(ctx context.Context, rfidUID, deviceID string, fails int, distanceCm float32, lightLevel int, classification int, confidence float32, recommendation string) (*MFARequest, error) {
	req := &MFARequest{
		RfidUID:        rfidUID,
		DeviceID:       deviceID,
		Fails:          fails,
		DistanceCm:     distanceCm,
		LightLevel:     lightLevel,
		Classification: classification,
		Confidence:     confidence,
		Recommendation: recommendation,
		Status:         "pending",
	}

	err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to save mfa request to db: %w", err)
	}

	// Fetch user name if user exists
	u, err := s.userService.GetUserByUID(ctx, rfidUID)
	if err == nil && u != nil && u.Name != nil {
		req.UserName = u.Name
	}

	// Broadcast websocket notification
	if s.wsHub != nil {
		s.wsHub.Broadcast(map[string]interface{}{
			"type": "mfa_request",
			"data": req,
		})
	}

	return req, nil
}

func (s *Service) ListRequests(ctx context.Context) ([]MFARequest, error) {
	return s.repo.ListAll(ctx)
}

func (s *Service) ApproveRequest(ctx context.Context, id int) error {
	req, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if req == nil {
		return fmt.Errorf("request not found: %d", id)
	}
	if req.Status != "pending" {
		return fmt.Errorf("request is already %s", req.Status)
	}

	err = s.repo.UpdateStatus(ctx, id, "approved")
	if err != nil {
		return err
	}

	// Publish unlock message to MQTT
	if s.mqttClient != nil {
		go func() {
			if err := s.mqttClient.PublishOpenDoor(); err != nil {
				log.Printf("Failed to publish open door command to MQTT in MFA approval: %v\n", err)
			}
		}()
	}

	// Broadcast WS state change
	if s.wsHub != nil {
		s.wsHub.Broadcast(map[string]interface{}{
			"type": "mfa_approved",
			"id":   id,
		})
	}

	return nil
}

func (s *Service) RejectRequest(ctx context.Context, id int) error {
	req, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if req == nil {
		return fmt.Errorf("request not found: %d", id)
	}
	if req.Status != "pending" {
		return fmt.Errorf("request is already %s", req.Status)
	}

	err = s.repo.UpdateStatus(ctx, id, "rejected")
	if err != nil {
		return err
	}

	// Block the user card
	isBlocked := true
	_, err = s.userService.UpdateUser(ctx, req.RfidUID, nil, nil, nil, &isBlocked)
	if err != nil {
		log.Printf("Failed to block user RFID card %s on MFA rejection: %v\n", req.RfidUID, err)
	}

	// Broadcast WS state change
	if s.wsHub != nil {
		s.wsHub.Broadcast(map[string]interface{}{
			"type": "mfa_rejected",
			"id":   id,
		})
	}

	return nil
}
