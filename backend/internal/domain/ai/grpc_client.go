package ai

import (
	"context"
	"fmt"

	smartlock "github.com/rubenalves-dev/smart-lock-distributed-system/internal/gen"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/models"
)

type grpcClient struct {
	client smartlock.AIServiceClient
}

func NewGRPCClient(client smartlock.AIServiceClient) AIService {
	return &grpcClient{client: client}
}

func (g *grpcClient) PredictSeverity(ctx context.Context, event models.SensorPayload) (int32, float32, string, error) {
	if g.client == nil {
		return 0, 0, "", fmt.Errorf("gRPC AI Service client is nil")
	}

	req := &smartlock.PredictSeverityRequest{
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
	}

	resp, err := g.client.PredictSeverity(ctx, req)
	if err != nil {
		return 0, 0, "", err
	}

	return int32(resp.Classification), resp.Confidence, resp.Recommendation, nil
}

func (g *grpcClient) RetrainModel(ctx context.Context, epochs int32, datasetPath string) (bool, string, error) {
	if g.client == nil {
		return false, "gRPC AI Service client is nil", fmt.Errorf("gRPC AI Service client is nil")
	}

	req := &smartlock.RetrainModelRequest{
		Epochs:      epochs,
		DatasetPath: datasetPath,
	}

	resp, err := g.client.RetrainModel(ctx, req)
	if err != nil {
		return false, "", err
	}

	return resp.Success, resp.Message, nil
}
