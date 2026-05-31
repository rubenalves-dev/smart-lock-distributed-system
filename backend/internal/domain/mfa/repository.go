package mfa

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Repository interface {
	Create(ctx context.Context, req *MFARequest) error
	UpdateStatus(ctx context.Context, id int, status string) error
	FindByID(ctx context.Context, id int) (*MFARequest, error)
	ListAll(ctx context.Context) ([]MFARequest, error)
}

type sqlRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) Create(ctx context.Context, req *MFARequest) error {
	query := `
		INSERT INTO mfa_requests (rfid_uid, device_id, fails, distance_cm, light_level, classification, confidence, recommendation, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, created_at, updated_at`
	err := r.db.QueryRowContext(ctx, query,
		req.RfidUID, req.DeviceID, req.Fails, req.DistanceCm, req.LightLevel,
		req.Classification, req.Confidence, req.Recommendation, req.Status,
	).Scan(&req.ID, &req.CreatedAt, &req.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create mfa request: %w", err)
	}
	return nil
}

func (r *sqlRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	query := `
		UPDATE mfa_requests
		SET status = $1, updated_at = NOW()
		WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update mfa request status: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("mfa request not found: %d", id)
	}
	return nil
}

func (r *sqlRepository) FindByID(ctx context.Context, id int) (*MFARequest, error) {
	query := `
		SELECT r.id, r.rfid_uid, r.device_id, r.fails, r.distance_cm, r.light_level, r.classification, r.confidence, r.recommendation, r.status, r.created_at, r.updated_at, u.name
		FROM mfa_requests r
		LEFT JOIN users u ON r.rfid_uid = u.rfid_uid
		WHERE r.id = $1`
	var req MFARequest
	var userName sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&req.ID, &req.RfidUID, &req.DeviceID, &req.Fails, &req.DistanceCm, &req.LightLevel,
		&req.Classification, &req.Confidence, &req.Recommendation, &req.Status, &req.CreatedAt, &req.UpdatedAt,
		&userName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find mfa request by ID: %w", err)
	}
	if userName.Valid {
		req.UserName = &userName.String
	}
	return &req, nil
}

func (r *sqlRepository) ListAll(ctx context.Context) ([]MFARequest, error) {
	query := `
		SELECT r.id, r.rfid_uid, r.device_id, r.fails, r.distance_cm, r.light_level, r.classification, r.confidence, r.recommendation, r.status, r.created_at, r.updated_at, u.name
		FROM mfa_requests r
		LEFT JOIN users u ON r.rfid_uid = u.rfid_uid
		ORDER BY r.created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list mfa requests: %w", err)
	}
	defer rows.Close()

	requests := []MFARequest{}
	for rows.Next() {
		var req MFARequest
		var userName sql.NullString
		err := rows.Scan(
			&req.ID, &req.RfidUID, &req.DeviceID, &req.Fails, &req.DistanceCm, &req.LightLevel,
			&req.Classification, &req.Confidence, &req.Recommendation, &req.Status, &req.CreatedAt, &req.UpdatedAt,
			&userName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan mfa request row: %w", err)
		}
		if userName.Valid {
			req.UserName = &userName.String
		}
		requests = append(requests, req)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating mfa request rows: %w", err)
	}
	return requests, nil
}
