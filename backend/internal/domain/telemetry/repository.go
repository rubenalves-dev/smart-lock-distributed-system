package telemetry

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/models"
)

type Repository interface {
	Save(ctx context.Context, payload models.SensorPayload) error
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
