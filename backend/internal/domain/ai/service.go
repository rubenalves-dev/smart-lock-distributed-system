package ai

import (
	"context"
	"fmt"

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

func (s *service) RetrainModel(ctx context.Context, epochs int32, datasetPath string) (bool, string, error) {
	return s.grpcClient.RetrainModel(ctx, epochs, datasetPath)
}

func (s *service) EvaluateModel(ctx context.Context, datasetPath string) (*models.EvaluationResult, error) {
	fmt.Println("DEBUG: A entrar no EvaluateModel do service.go")
	result, err := s.grpcClient.EvaluateModel(ctx, datasetPath)
	if err != nil {
		return nil, err
	}

	eval := &Evaluation{
		DatasetPath: datasetPath,
		Accuracy:    float64(result.Metrics.Accuracy),
	}

	if err := s.repo.SaveEvaluation(ctx, eval); err != nil {
		fmt.Printf("ERRO: Falha ao guardar na base de dados: %v\n", err)
		return nil, err
	}

	return result, nil
}
