package main

import (
	"context"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/config"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/config/db"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
)

func main() {

	cfg, err := initConfig("config/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := db.NewConDB(ctx, cfg)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	if pool != nil {
		defer pool.Close()
	}

	app := internal.NewApp(ctx, cfg, pool)
	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run app: %v", err)
	}

}

func initConfig(pathYAML string) (*config.Config, error) {
	cfg, err := config.LoadYAML(pathYAML)
	if err != nil {
		return nil, err
	}

	flags, err := config.ParseFlags()
	if err != nil {
		return cfg, err
	}
	if flags.RunAddress != "" {
		cfg.Server.Address = flags.RunAddress
	}
	if flags.DatabaseURI != "" {
		cfg.Postgres.DSN = flags.DatabaseURI
	}
	if flags.AccrualSystemAddress != "" {
		cfg.Client.Address = flags.AccrualSystemAddress
	}

	env, err := config.LoadEnvVar()
	if err != nil {
		return cfg, err
	}
	if env.RunAddress != "" {
		cfg.Server.Address = env.RunAddress
	}
	if env.DatabaseURI != "" {
		cfg.Postgres.DSN = env.DatabaseURI
	}
	if env.AccrualSystemAddress != "" {
		cfg.Client.Address = env.AccrualSystemAddress
	}

	return cfg, nil
}
