package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Environment string `env:"ENV" env-default:"dev"` // dev, prod, test

	HTTP struct {
		Host           string        `env:"HTTP_HOST" env-default:"0.0.0.0"`
		Port           int           `env:"HTTP_PORT" env-default:"8080"`
		ReadTimeout    time.Duration `env:"HTTP_READ_TIMEOUT" env-default:"5s"`
		WriteTimeout   time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"10s"`
		IdleTimeout    time.Duration `env:"HTTP_IDLE_TIMEOUT" env-default:"120s"`
		MaxHeaderBytes int           `env:"HTTP_MAX_HEADER_BYTES" env-default:"1048576"` // 1MB
	}

	Postgres struct {
		Host     string `env:"POSTGRES_HOST" env-default:"localhost"`
		Port     int    `env:"POSTGRES_PORT" env-default:"5432"`
		User     string `env:"POSTGRES_USER" env-default:"postgres"`
		Password string `env:"POSTGRES_PASSWORD" env-default:"postgres"`
		DBName   string `env:"POSTGRES_DB" env-default:"pr_service"`
		SSLMode  string `env:"POSTGRES_SSLMODE" env-default:"disable"`
	}

	Logger struct {
		Level string `env:"LOGGER_LEVEL" env-default:"info"` // debug, info, warn, error
	}

	PRService struct {
		MaxReviewers int `env:"MAX_REVIEWERS" env-default:"2"`
	}
}

func ParseConfig(path string) (*Config, error) {
	cfg := &Config{}

	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}
