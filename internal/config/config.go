package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DataBaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func GetConfig() Config {
	cfg := Config{}

	flag.StringVar(&cfg.RunAddress, "a", ":8080", "Server address")
	flag.StringVar(&cfg.DataBaseURI, "d", "", "DB connect URI")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "http://127.0.0.1", "Accrual system address")
	flag.Parse()

	env.Parse(&cfg)

	return cfg
}
