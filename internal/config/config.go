package config

import (
	"log"
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	RunAddress           string        `env:"RUN_ADDRESS"`
	DataBaseURI          string        `env:"DATABASE_URI"`
	AccrualSystemAddress string        `env:"ACCRUAL_SYSTEM_ADDRESS"`
	AwaitTime            time.Duration `env:"AWAIT_TIME"`
	WorkersCount         int           `env:"WORKERS_COUNT"`
	SecretKeyRaw         string        `env:"SECRET_KEY"`
	SecretKey            []byte
}

func GetConfig() Config {
	cfg := Config{}

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal("Failed parse env config:", err)
		return cfg
	}

	cfg.SecretKey = []byte(cfg.SecretKeyRaw)

	return cfg
}
