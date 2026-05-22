package config

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	RabbitMQURL   string `env:"RABBITMQ_URL" envDefault:"amqp://guest:guest@localhost:5672/"`
	AIServiceAddr string `env:"AI_SERVICE_ADDR" envDefault:"localhost:50051"`
	Port          string `env:"PORT" envDefault:"8080"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
