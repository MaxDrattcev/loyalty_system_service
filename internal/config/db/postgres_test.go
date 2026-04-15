package db

import (
	"context"
	"testing"

	"github.com/MaxDrattcev/loyalty_system_service.git/internal/config"
	"github.com/stretchr/testify/require"
)

func TestNewConDB_EarlyValidation_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr string
	}{
		{
			name: "empty dsn",
			cfg: &config.Config{
				Postgres: config.PostgresConfig{
					DSN: "",
				},
			},
			wantErr: "database DSN is empty",
		},
		{
			name: "invalid dsn parse error",
			cfg: &config.Config{
				Postgres: config.PostgresConfig{
					DSN: ":// definitely-not-a-dsn",
				},
			},
			wantErr: "parse dsn:",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pool, err := NewConDB(context.Background(), tt.cfg)

			require.Error(t, err)
			require.Nil(t, pool)
			require.Contains(t, err.Error(), tt.wantErr)
		})
	}
}
