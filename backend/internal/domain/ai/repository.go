package ai

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Evaluation struct {
	ID          int       `json:"id"`
	DatasetPath string    `json:"dataset_path"`
	Accuracy    float64   `json:"accuracy"`
	CreatedAt   time.Time `json:"created_at"`
}

type Retrain struct {
	ID                 int       `json:"id"`
	DatasetPath        string    `json:"dataset_path"`
	Epochs             int       `json:"epochs"`
	Success            bool      `json:"success"`
	Message            string    `json:"message"`
	TrainAccuracy      float64   `json:"train_accuracy"`
	ValidationAccuracy float64   `json:"validation_accuracy"`
	TrainLoss          float64   `json:"train_loss"`
	ValidationLoss     float64   `json:"validation_loss"`
	Underfitting       bool      `json:"underfitting_detected"`
	Overfitting        bool      `json:"overfitting_detected"`
	CreatedAt          time.Time `json:"created_at"`
}

type sqlRepository struct {
	db *sql.DB
}

// NewRepository cria a instância que satisfaz a interface Repository
func NewRepository(db *sql.DB) Repository {
	return &sqlRepository{db: db}
}

// SaveEvaluation implementa o método definido na interface
func (r *sqlRepository) SaveEvaluation(ctx context.Context, e *Evaluation) error {
	query := `INSERT INTO ai_evaluations (dataset_path, accuracy, created_at) VALUES ($1, $2, NOW())`

	fmt.Println("Tentando gravar na base de dados...")

	_, err := r.db.ExecContext(ctx, query, e.DatasetPath, e.Accuracy)

	if err != nil {
		fmt.Printf("Erro ao gravar na base de dados: %v\n", err)
		return err
	}

	fmt.Println("Gravado com sucesso!")
	return nil
}

func (r *sqlRepository) ListEvaluations(ctx context.Context) ([]Evaluation, error) {
	query := `SELECT id, dataset_path, accuracy, created_at FROM ai_evaluations ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	evals := []Evaluation{}
	for rows.Next() {
		var e Evaluation
		if err := rows.Scan(&e.ID, &e.DatasetPath, &e.Accuracy, &e.CreatedAt); err != nil {
			return nil, err
		}
		evals = append(evals, e)
	}
	return evals, nil
}

func (r *sqlRepository) SaveRetrain(ctx context.Context, rt *Retrain) error {
	query := `
		INSERT INTO ai_retrains (
			dataset_path, epochs, success, message, train_accuracy, validation_accuracy, 
			train_loss, validation_loss, underfitting_detected, overfitting_detected, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())`

	_, err := r.db.ExecContext(ctx, query,
		rt.DatasetPath, rt.Epochs, rt.Success, rt.Message, rt.TrainAccuracy, rt.ValidationAccuracy,
		rt.TrainLoss, rt.ValidationLoss, rt.Underfitting, rt.Overfitting,
	)
	return err
}

func (r *sqlRepository) ListRetrains(ctx context.Context) ([]Retrain, error) {
	query := `
		SELECT 
			id, dataset_path, epochs, success, COALESCE(message, ''), COALESCE(train_accuracy, 0.0), COALESCE(validation_accuracy, 0.0), 
			COALESCE(train_loss, 0.0), COALESCE(validation_loss, 0.0), COALESCE(underfitting_detected, false), COALESCE(overfitting_detected, false), created_at 
		FROM ai_retrains 
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	retrains := []Retrain{}
	for rows.Next() {
		var rt Retrain
		err := rows.Scan(
			&rt.ID, &rt.DatasetPath, &rt.Epochs, &rt.Success, &rt.Message, &rt.TrainAccuracy, &rt.ValidationAccuracy,
			&rt.TrainLoss, &rt.ValidationLoss, &rt.Underfitting, &rt.Overfitting, &rt.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		retrains = append(retrains, rt)
	}
	return retrains, nil
}
