package config

import (
	"flag"
	"github.com/aseptimu/internal/utils"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"log/slog"
)

type ConfigType struct {
	ServerAddress        string `env:"RUN_ADDRESS"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DSN                  string `env:"DATABASE_URI"`
	SecretKey            string `env:"SECRET_KEY"`
}

func NewConfig() *ConfigType {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("Error loading .env file")
	}
	config := &ConfigType{}

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&config.AccrualSystemAddress, "r", "localhost:8085", "accrual system address")
	flag.StringVar(&config.DSN, "d", "", "PostgreSQL connection DSN")
	flag.StringVar(&config.SecretKey, "s", "", "Secret key")

	flag.Parse()

	if err := env.Parse(config); err != nil {
		slog.Error("Error parsing config", "error", err)
	}

	if config.SecretKey == "" {
		config.SecretKey = utils.GenerateRandomSecretKey()
	}

	return config
}
