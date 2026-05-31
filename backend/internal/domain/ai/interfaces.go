package ai

import (
	"context"

	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/models"
)

type AIService interface {
	PredictSeverity(ctx context.Context, event models.SensorPayload) (int32, float32, string, error)
	RetrainModel(ctx context.Context, epochs int32, datasetPath string) (*models.RetrainResult, error)
	EvaluateModel(ctx context.Context, datasetPath string) (*models.EvaluationResult, error)
}

type GRPCClient interface {
	PredictSeverity(ctx context.Context, event models.SensorPayload) (int32, float32, string, error)
	RetrainModel(ctx context.Context, epochs int32, datasetPath string) (*models.RetrainResult, error)
	EvaluateModel(ctx context.Context, datasetPath string) (*models.EvaluationResult, error)
}

type Repository interface {
	SaveEvaluation(ctx context.Context, e *Evaluation) error
}
