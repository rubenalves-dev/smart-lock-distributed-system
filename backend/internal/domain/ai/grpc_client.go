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

func (g *grpcClient) RetrainModel(ctx context.Context, epochs int32, datasetPath string) (*models.RetrainResult, error) {
	if g.client == nil {
		return nil, fmt.Errorf("gRPC AI Service client is nil")
	}

	req := &smartlock.RetrainModelRequest{
		Epochs:      epochs,
		DatasetPath: datasetPath,
	}

	resp, err := g.client.RetrainModel(ctx, req)
	if err != nil {
		return nil, err
	}

	var diagnostics *models.TrainingDiagnostics
	if resp.Diagnostics != nil {
		diagnostics = &models.TrainingDiagnostics{
			TrainAccuracy:        resp.Diagnostics.TrainAccuracy,
			ValidationAccuracy:   resp.Diagnostics.ValidationAccuracy,
			TrainLoss:            resp.Diagnostics.TrainLoss,
			ValidationLoss:       resp.Diagnostics.ValidationLoss,
			UnderfittingDetected: resp.Diagnostics.UnderfittingDetected,
			OverfittingDetected:  resp.Diagnostics.OverfittingDetected,
		}
	}

	return &models.RetrainResult{
		Success:     resp.Success,
		Message:     resp.Message,
		Diagnostics: diagnostics,
	}, nil
}

func (g *grpcClient) EvaluateModel(ctx context.Context, datasetPath string) (*models.EvaluationResult, error) {
	if g.client == nil {
		return nil, fmt.Errorf("gRPC AI Service client is nil")
	}

	req := &smartlock.EvaluateModelRequest{
		DatasetPath: datasetPath,
	}

	resp, err := g.client.EvaluateModel(ctx, req)
	if err != nil {
		return nil, err
	}

	matrix := make([][]int32, len(resp.ConfusionMatrix))
	for i, row := range resp.ConfusionMatrix {
		matrix[i] = row.Values
	}

	var metrics models.EvaluationMetrics
	if resp.Metrics != nil {
		metrics = models.EvaluationMetrics{
			Accuracy:       resp.Metrics.Accuracy,
			PrecisionMacro: resp.Metrics.PrecisionMacro,
			RecallMacro:    resp.Metrics.RecallMacro,
			F1Macro:        resp.Metrics.F1Macro,
		}
	}

	var binaryMetrics models.BinaryEvaluationMetrics
	if resp.BinaryMetrics != nil {
		binaryMetrics = models.BinaryEvaluationMetrics{
			Accuracy:  resp.BinaryMetrics.Accuracy,
			Precision: resp.BinaryMetrics.Precision,
			Recall:    resp.BinaryMetrics.Recall,
			F1:        resp.BinaryMetrics.F1,
		}
	}

	return &models.EvaluationResult{
		ConfusionMatrix: matrix,
		Metrics:         metrics,
		BinaryMetrics:   binaryMetrics,
	}, nil
}

