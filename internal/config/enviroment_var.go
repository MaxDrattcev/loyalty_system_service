package config

import (
	"fmt"
	"github.com/caarlos0/env/v6"
)

// EnvVar contains supported environment variable overrides.
type EnvVar struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

// LoadEnvVar reads environment variables into EnvVar structure.
func LoadEnvVar() (EnvVar, error) {
	var envVar EnvVar
	if err := env.Parse(&envVar); err != nil {
		return EnvVar{}, fmt.Errorf("error parsing environment variables: %w", err)
	}
	return envVar, nil
}
