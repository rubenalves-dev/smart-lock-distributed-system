package core

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type PostgresClient struct {
	DB *sql.DB
}

func NewPostgresClient(url string) (*PostgresClient, error) {
	var db *sql.DB
	var err error

	// Retry connecting to PostgreSQL
	for i := range 15 {
		db, err = sql.Open("postgres", url)
		if err == nil {
			err = db.Ping()
			if err == nil {
				log.Println("Connected to PostgreSQL DB successfully")
				client := &PostgresClient{DB: db}
				if err := client.initSchema(); err != nil {
					db.Close()
					return nil, fmt.Errorf("failed to initialize schema: %w", err)
				}
				return client, nil
			}
		}
		log.Printf("Failed to connect to PostgreSQL, retrying in 2 seconds... (%d/15): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to PostgreSQL after retries: %w", err)
}

func (p *PostgresClient) initSchema() error {
	usersTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		rfid_uid VARCHAR(50) UNIQUE NOT NULL,
		name VARCHAR(100),
		email VARCHAR(100),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	telemetryTableQuery := `
	CREATE TABLE IF NOT EXISTS telemetry (
		id SERIAL PRIMARY KEY,
		device_id VARCHAR(50) NOT NULL,
		event VARCHAR(50) NOT NULL,
		details TEXT,
		status VARCHAR(50),
		distance_cm REAL,
		light_level INT,
		fails INT,
		rfid_uid VARCHAR(50),
		rssi INT,
		uptime REAL,
		timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := p.DB.Exec(usersTableQuery); err != nil {
		return fmt.Errorf("failed creating users table: %w", err)
	}

	if _, err := p.DB.Exec(telemetryTableQuery); err != nil {
		return fmt.Errorf("failed creating telemetry table: %w", err)
	}

	log.Println("PostgreSQL schema initialized successfully")
	return nil
}

func (p *PostgresClient) Ping(ctx context.Context) error {
	if p.DB == nil {
		return fmt.Errorf("database connection is nil")
	}
	return p.DB.PingContext(ctx)
}

func (p *PostgresClient) Close() error {
	if p.DB != nil {
		return p.DB.Close()
	}
	return nil
}
