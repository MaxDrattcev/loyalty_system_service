package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeTestConfig(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	content := `
server:
  address: "yaml-address:8080"
postgres:
  host: "localhost"
  port: "5432"
  user: "yaml_user"
  password: "yaml_pass"
  db_name: "yaml_db"
  ssl_mode: "disable"
  max_conns: 10
  min_conns: 1
  max_conn_lifetime: 30
  max_conn_idle_time: 5
  path_migration: "file://migrations"
jwt_token:
  secret: "yaml-secret"
  expires_at: 60
client:
  address: "yaml-accrual:8081"
worker:
  queue_size: 100
  worker_count: 4
  interval_get_accrual: 10
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func resetFlagsAndArgs(t *testing.T, args []string) {
	t.Helper()

	oldArgs := os.Args
	oldCommandLine := flag.CommandLine

	os.Args = args
	flag.CommandLine = flag.NewFlagSet(args[0], flag.ContinueOnError)

	t.Cleanup(func() {
		os.Args = oldArgs
		flag.CommandLine = oldCommandLine
	})
}

func TestInitConfig_LoadYAMLFail(t *testing.T) {
	resetFlagsAndArgs(t, []string{"cmd"})

	cfg, err := initConfig("definitely-not-exists.yaml")
	require.Error(t, err)
	require.Nil(t, cfg)
	require.Contains(t, err.Error(), "failed reading config file")
}

func TestInitConfig_OverridePriority_YAML_Flags_Env(t *testing.T) {
	cfgPath := writeTestConfig(t)

	t.Setenv("RUN_ADDRESS", "")
	t.Setenv("DATABASE_URI", "")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "")

	resetFlagsAndArgs(t, []string{
		"cmd",
		"-a=flag-address:9000",
		"-d=postgres://flag_user:flag_pass@localhost:5432/flag_db?sslmode=disable",
		"-r=flag-accrual:9001",
	})

	t.Setenv("RUN_ADDRESS", "env-address:10000")
	t.Setenv("DATABASE_URI", "postgres://env_user:env_pass@localhost:5432/env_db?sslmode=disable")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "env-accrual:10001")

	cfg, err := initConfig(cfgPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.Equal(t, "env-address:10000", cfg.Server.Address)
	require.Equal(t, "postgres://env_user:env_pass@localhost:5432/env_db?sslmode=disable", cfg.Postgres.DSN)
	require.Equal(t, "env-accrual:10001", cfg.Client.Address)
}

func TestInitConfig_ReturnsConfigWhenParseFlagsFails(t *testing.T) {
	cfgPath := writeTestConfig(t)

	resetFlagsAndArgs(t, []string{"cmd", "-unknown=1"})

	cfg, err := initConfig(cfgPath)
	require.Error(t, err)
	require.NotNil(t, cfg)
	require.Contains(t, err.Error(), "unknown flag")
	require.Equal(t, "yaml-address:8080", cfg.Server.Address)
}
