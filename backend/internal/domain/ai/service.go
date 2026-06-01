package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/models"
)

type service struct {
	repo       Repository
	grpcClient GRPCClient
}

func NewService(repo Repository, grpcClient GRPCClient) AIService {
	return &service{
		repo:       repo,
		grpcClient: grpcClient,
	}
}

func (s *service) PredictSeverity(ctx context.Context, event models.SensorPayload) (int32, float32, string, error) {
	return s.grpcClient.PredictSeverity(ctx, event)
}

func (s *service) RetrainModel(ctx context.Context, epochs int32, datasetPath string) (*models.RetrainResult, error) {
	result, err := s.grpcClient.RetrainModel(ctx, epochs, datasetPath)
	if err != nil {
		return nil, err
	}

	label := datasetPath
	if datasetPath == "database" {
		label = "[Database Telemetry]"
	} else if len(datasetPath) > 200 || strings.Contains(datasetPath, "\n") || strings.HasPrefix(datasetPath, "fails,distance_cm") {
		label = "[Uploaded CSV]"
	}

	retrainObj := &Retrain{
		DatasetPath: label,
		Epochs:      int(epochs),
		Success:     result.Success,
		Message:     result.Message,
	}
	if result.Diagnostics != nil {
		retrainObj.TrainAccuracy = float64(result.Diagnostics.TrainAccuracy)
		retrainObj.ValidationAccuracy = float64(result.Diagnostics.ValidationAccuracy)
		retrainObj.TrainLoss = float64(result.Diagnostics.TrainLoss)
		retrainObj.ValidationLoss = float64(result.Diagnostics.ValidationLoss)
		retrainObj.Underfitting = result.Diagnostics.UnderfittingDetected
		retrainObj.Overfitting = result.Diagnostics.OverfittingDetected
	}

	if err := s.repo.SaveRetrain(ctx, retrainObj); err != nil {
		fmt.Printf("ERRO: Falha ao guardar retreino na base de dados: %v\n", err)
	}

	return result, nil
}

func (s *service) EvaluateModel(ctx context.Context, datasetPath string) (*models.EvaluationResult, error) {
	fmt.Println("DEBUG: A entrar no EvaluateModel do service.go")
	result, err := s.grpcClient.EvaluateModel(ctx, datasetPath)
	if err != nil {
		return nil, err
	}

	label := datasetPath
	if len(datasetPath) > 200 || strings.Contains(datasetPath, "\n") || strings.HasPrefix(datasetPath, "fails,distance_cm") {
		label = "[Uploaded CSV]"
	}

	eval := &Evaluation{
		DatasetPath: label,
		Accuracy:    float64(result.Metrics.Accuracy),
	}

	if err := s.repo.SaveEvaluation(ctx, eval); err != nil {
		fmt.Printf("ERRO: Falha ao guardar na base de dados: %v\n", err)
		return nil, err
	}

	return result, nil
}

func (s *service) ListEvaluations(ctx context.Context) ([]Evaluation, error) {
	return s.repo.ListEvaluations(ctx)
}

func (s *service) ListRetrains(ctx context.Context) ([]Retrain, error) {
	return s.repo.ListRetrains(ctx)
}
