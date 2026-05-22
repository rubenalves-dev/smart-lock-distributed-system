package ai

import (
	"context"

	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/models"
)

type AIService interface {
	PredictSeverity(ctx context.Context, event models.SensorPayload) (int32, float32, string, error)
	RetrainModel(ctx context.Context, epochs int32, datasetPath string) (bool, string, error)
}
