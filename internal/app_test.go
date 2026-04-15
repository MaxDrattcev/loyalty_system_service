package internal

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/MaxDrattcev/loyalty_system_service.git/internal/config"
	"github.com/stretchr/testify/require"
)

func testConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Address: "127.0.0.1:0",
		},
		Postgres: config.PostgresConfig{
			DSN:           "postgres://user:pass@localhost:5432/db?sslmode=disable",
			PathMigration: "file://migrations",
		},
		JWTToken: config.JWTTokenConfig{
			Secret:    "test-secret",
			ExpiresAt: 60,
		},
		Client: config.ClientConfig{
			Address: "127.0.0.1:8081",
		},
		Worker: config.WorkerConfig{
			QueueSize:          1,
			WorkerCount:        0,
			IntervalGetAccrual: 1,
		},
	}
}

func TestNewApp_ReturnsInitializedApp(t *testing.T) {
	cfg := testConfig()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	app := NewApp(ctx, cfg, nil)

	require.NotNil(t, app)
	require.NotNil(t, app.router)
	require.NotNil(t, app.userHandler)
	require.Equal(t, cfg, app.cfg)
}

func TestApp_Run_ReturnsErrorOnInvalidAddress(t *testing.T) {
	app := &App{
		cfg: &config.Config{
			Server: config.ServerConfig{
				Address: "bad-address",
			},
		},
		router: http.NewServeMux(),
	}

	done := make(chan error, 1)
	go func() {
		done <- app.Run()
	}()

	select {
	case err := <-done:
		require.Error(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not return in time")
	}
}
