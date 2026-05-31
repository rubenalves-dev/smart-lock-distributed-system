package ai

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Evaluation struct {
	ID          int
	DatasetPath string
	Accuracy    float64
	CreatedAt   time.Time
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
