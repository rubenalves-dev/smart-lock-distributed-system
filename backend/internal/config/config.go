package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port           int    `env:"PORT" envDefault:"8080"`
	PostgresURL    string `env:"POSTGRES_URL" envDefault:"postgres://user:password@postgres:5432/lock_db?sslmode=disable"`
	MQTTBroker     string `env:"MQTT_BROKER" envDefault:"mosquitto_broker"`
	InfluxDBURL    string `env:"INFLUXDB_URL" envDefault:"http://influxdb:8086"`
	InfluxDBToken  string `env:"INFLUXDB_TOKEN" envDefault:"ZYPGu_Lu6NaP8M4iT5_TLx1xSZag9sAbR9i2vH8zr7P253VcxIMuXHbbkYagn2bOzfVZCRpsrIH5_77r3G1Mag=="`
	InfluxDBOrg    string `env:"INFLUXDB_ORG" envDefault:"A MinhaOrganizacao"`
	InfluxDBBucket string `env:"INFLUXDB_BUCKET" envDefault:"esp32_pings"`
	RabbitMQURL    string `env:"RABBITMQ_URL" envDefault:"amqp://guest:guest@rabbitmq:5672/"`
	AIServiceAddr  string `env:"AI_SERVICE_ADDR" envDefault:"ai-service:50051"`
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config env vars: %w", err)
	}
	return cfg, nil
}
