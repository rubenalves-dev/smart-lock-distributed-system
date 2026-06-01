package telemetry

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/models"
)

type Repository interface {
	Save(ctx context.Context, payload models.SensorPayload) error
	GetLatest(ctx context.Context, deviceID string) (*models.SensorPayload, error)
	GetAll(ctx context.Context) ([]models.SensorPayload, error)
	GetDevices(ctx context.Context) ([]string, error)
}

type sqlRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) Save(ctx context.Context, p models.SensorPayload) error {
	query := `
		INSERT INTO telemetry (device_id, event, details, status, distance_cm, light_level, fails, rfid_uid, rssi, uptime, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())`

	_, err := r.db.ExecContext(ctx, query,
		p.DeviceID,
		p.Event,
		p.Details,
		p.Status,
		p.DistanceCm,
		p.LightLevel,
		p.Fails,
		p.RfidUID,
		p.RSSI,
		p.Uptime,
	)
	if err != nil {
		return fmt.Errorf("failed to save telemetry: %w", err)
	}
	return nil
}

func (r *sqlRepository) GetLatest(ctx context.Context, deviceID string) (*models.SensorPayload, error) {
	query := `
		SELECT 
			device_id, 
			event, 
			COALESCE(details, ''), 
			COALESCE(status, ''), 
			COALESCE(distance_cm, 0.0), 
			COALESCE(light_level, 0), 
			COALESCE(fails, 0), 
			COALESCE(rfid_uid, ''), 
			COALESCE(rssi, 0), 
			COALESCE(uptime, 0.0)
		FROM telemetry
		WHERE device_id = $1
		ORDER BY timestamp DESC
		LIMIT 1`

	var p models.SensorPayload
	err := r.db.QueryRowContext(ctx, query, deviceID).Scan(
		&p.DeviceID,
		&p.Event,
		&p.Details,
		&p.Status,
		&p.DistanceCm,
		&p.LightLevel,
		&p.Fails,
		&p.RfidUID,
		&p.RSSI,
		&p.Uptime,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest telemetry: %w", err)
	}
	return &p, nil
}

func (r *sqlRepository) GetAll(ctx context.Context) ([]models.SensorPayload, error) {
	query := `
		SELECT 
			device_id, 
			event, 
			COALESCE(details, ''), 
			COALESCE(status, ''), 
			COALESCE(distance_cm, 0.0), 
			COALESCE(light_level, 0), 
			COALESCE(fails, 0), 
			COALESCE(rfid_uid, ''), 
			COALESCE(rssi, 0), 
			COALESCE(uptime, 0.0)
		FROM telemetry
		ORDER BY timestamp DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all telemetry: %w", err)
	}
	defer rows.Close()

	payloads := []models.SensorPayload{}
	for rows.Next() {
		var p models.SensorPayload
		err := rows.Scan(
			&p.DeviceID,
			&p.Event,
			&p.Details,
			&p.Status,
			&p.DistanceCm,
			&p.LightLevel,
			&p.Fails,
			&p.RfidUID,
			&p.RSSI,
			&p.Uptime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan telemetry row: %w", err)
		}
		payloads = append(payloads, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating telemetry rows: %w", err)
	}
	return payloads, nil
}

func (r *sqlRepository) GetDevices(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT device_id FROM telemetry`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices: %w", err)
	}
	defer rows.Close()

	devices := []string{}
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}
		devices = append(devices, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating devices: %w", err)
	}
	return devices, nil
}

