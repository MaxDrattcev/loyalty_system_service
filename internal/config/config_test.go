package config

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadYAML_TableDriven(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir := t.TempDir()
		cfgPath := filepath.Join(dir, "config.yaml")

		content := `
server:
  address: "localhost:8080"
postgres:
  host: "localhost"
  port: "5432"
  user: "postgres"
  password: "secret"
  db_name: "gophermart"
  ssl_mode: "disable"
  max_conns: 10
  min_conns: 1
  max_conn_lifetime: 30
  max_conn_idle_time: 5
  path_migration: "file://migrations"
jwt_token:
  secret: "jwt-secret"
  expires_at: 60
client:
  address: "localhost:8081"
worker:
  queue_size: 100
  worker_count: 4
  interval_get_accrual: 10
`
		require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0o644))

		cfg, err := LoadYAML(cfgPath)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		require.Equal(t, "localhost:8080", cfg.Server.Address)
		require.Equal(t, "localhost", cfg.Postgres.Host)
		require.Equal(t, "5432", cfg.Postgres.Port)
		require.Equal(t, "postgres", cfg.Postgres.User)
		require.Equal(t, "secret", cfg.Postgres.Password)
		require.Equal(t, "gophermart", cfg.Postgres.DBName)
		require.Equal(t, "disable", cfg.Postgres.SSLMode)
		require.Equal(t, "postgres://postgres:secret@localhost:5432/gophermart?sslmode=disable", cfg.Postgres.DSN)
	})

	t.Run("file not found", func(t *testing.T) {
		cfg, err := LoadYAML("definitely-not-existing.yaml")
		require.Error(t, err)
		require.Nil(t, cfg)
		require.Contains(t, err.Error(), "failed reading config file")
	})

	t.Run("invalid yaml", func(t *testing.T) {
		dir := t.TempDir()
		cfgPath := filepath.Join(dir, "bad.yaml")
		require.NoError(t, os.WriteFile(cfgPath, []byte("server: ["), 0o644))

		cfg, err := LoadYAML(cfgPath)
		require.Error(t, err)
		require.Nil(t, cfg)
		require.Contains(t, err.Error(), "failed parsing config file")
	})
}

func TestInitDSN_TableDriven(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "with ssl mode",
			cfg: Config{
				Postgres: PostgresConfig{
					Host:     "localhost",
					Port:     "5432",
					User:     "postgres",
					Password: "secret",
					DBName:   "db",
					SSLMode:  "disable",
				},
			},
			want: "postgres://postgres:secret@localhost:5432/db?sslmode=disable",
		},
		{
			name: "without ssl mode",
			cfg: Config{
				Postgres: PostgresConfig{
					Host:     "localhost",
					Port:     "5432",
					User:     "postgres",
					Password: "secret",
					DBName:   "db",
				},
			},
			want: "postgres://postgres:secret@localhost:5432/db",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := initDSN(tt.cfg)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestLoadEnvVar(t *testing.T) {
	t.Setenv("RUN_ADDRESS", "localhost:8080")
	t.Setenv("DATABASE_URI", "postgres://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "localhost:8081")

	envVars, err := LoadEnvVar()
	require.NoError(t, err)

	require.Equal(t, "localhost:8080", envVars.RunAddress)
	require.Equal(t, "postgres://user:pass@localhost:5432/db?sslmode=disable", envVars.DatabaseURI)
	require.Equal(t, "localhost:8081", envVars.AccrualSystemAddress)
}

func resetFlagsForTest(t *testing.T, args []string) {
	t.Helper()

	oldArgs := os.Args
	oldFlagSet := flag.CommandLine

	os.Args = args
	flag.CommandLine = flag.NewFlagSet(args[0], flag.ContinueOnError)

	t.Cleanup(func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlagSet
	})
}

func TestParseFlags_TableDriven(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		errContain string
		want       *Flags
	}{
		{
			name: "success with all flags",
			args: []string{
				"cmd",
				"-a=localhost:8080",
				"-d=postgres://user:pass@localhost:5432/db?sslmode=disable",
				"-r=localhost:8081",
			},
			want: &Flags{
				RunAddress:           "localhost:8080",
				DatabaseURI:          "postgres://user:pass@localhost:5432/db?sslmode=disable",
				AccrualSystemAddress: "localhost:8081",
			},
		},
		{
			name:       "unknown flag",
			args:       []string{"cmd", "-x=1"},
			wantErr:    true,
			errContain: "unknown flag: -x",
		},
		{
			name: "no flags is valid",
			args: []string{"cmd"},
			want: &Flags{
				RunAddress:           "",
				DatabaseURI:          "",
				AccrualSystemAddress: "",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			resetFlagsForTest(t, tt.args)

			got, err := ParseFlags()

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
				require.Contains(t, err.Error(), tt.errContain)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCheckUnknownFlags(t *testing.T) {
	t.Run("known flags only", func(t *testing.T) {
		resetFlagsForTest(t, []string{"cmd", "-a=localhost:8080", "-d=db", "-r=acc"})

		_, err := ParseFlags()
		require.NoError(t, err)

		err = checkUnknownFlags()
		require.NoError(t, err)
	})

	t.Run("detect unknown long flag", func(t *testing.T) {
		resetFlagsForTest(t, []string{"cmd", "--not-known=1"})

		flag.String("a", "", "")
		flag.String("d", "", "")
		flag.String("r", "", "")

		err := checkUnknownFlags()
		require.Error(t, err)
		require.Contains(t, err.Error(), "unknown flag: -not-known")
	})
}
