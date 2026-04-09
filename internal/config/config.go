package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"net"
	"net/url"
	"os"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Postgres PostgresConfig `yaml:"postgres"`
	JWTToken JWTTokenConfig `yaml:"jwt_token"`
	Client   ClientConfig   `yaml:"client"`
	Worker   WorkerConfig   `yaml:"worker"`
}

type ServerConfig struct {
	Address string `yaml:"address"`
}
type PostgresConfig struct {
	Host            string `yaml:"host"`
	Port            string `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	DBName          string `yaml:"db_name"`
	SSLMode         string `yaml:"ssl_mode"`
	DSN             string
	MaxConns        int32  `yaml:"max_conns"`
	MinConns        int32  `yaml:"min_conns"`
	MaxConnLifetime int32  `yaml:"max_conn_lifetime"`
	MaxConnIdleTime int32  `yaml:"max_conn_idle_time"`
	PathMigration   string `yaml:"path_migration"`
}

type JWTTokenConfig struct {
	Secret    string `yaml:"secret"`
	ExpiresAt int32  `yaml:"expires_at"`
}

type ClientConfig struct {
	Address string `yaml:"address"`
}

type WorkerConfig struct {
	QueueSize          int   `yaml:"queue_size"`
	WorkerCount        int   `yaml:"worker_count"`
	IntervalGetAccrual int32 `yaml:"interval_get_accrual"`
}

func LoadYAML(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "config/config.yaml"
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed reading config file: %w", err)
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed parsing config file: %w", err)
	}
	config.Postgres.DSN = initDSN(config)
	return &config, nil
}

func initDSN(cfg Config) string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.Postgres.User, cfg.Postgres.Password),
		Host:   net.JoinHostPort(cfg.Postgres.Host, cfg.Postgres.Port),
		Path:   "/" + cfg.Postgres.DBName,
	}
	q := url.Values{}
	if cfg.Postgres.SSLMode != "" {
		q.Set("sslmode", cfg.Postgres.SSLMode)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
