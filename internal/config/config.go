package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	Environment string
	HTTPPort    string

	MySQL struct {
		DSN             string
		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime time.Duration
	}

	JWT struct {
		Secret         string
		AccessTokenTTL time.Duration
	}
}

func Load() (*Config, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvPrefix("TASK")

	// Defaults
	v.SetDefault("ENVIRONMENT", "dev")
	v.SetDefault("HTTP_PORT", "8080")
	v.SetDefault("MYSQL_MAX_OPEN_CONNS", 50)
	v.SetDefault("MYSQL_MAX_IDLE_CONNS", 25)
	v.SetDefault("MYSQL_CONN_MAX_LIFETIME", "30m")
	v.SetDefault("JWT_ACCESS_TOKEN_TTL", "1h")

	cfg := &Config{}
	cfg.Environment = v.GetString("ENVIRONMENT")
	cfg.HTTPPort = v.GetString("HTTP_PORT")
	cfg.MySQL.DSN = v.GetString("MYSQL_DSN")
	cfg.MySQL.MaxOpenConns = v.GetInt("MYSQL_MAX_OPEN_CONNS")
	cfg.MySQL.MaxIdleConns = v.GetInt("MYSQL_MAX_IDLE_CONNS")
	cfg.MySQL.ConnMaxLifetime = v.GetDuration("MYSQL_CONN_MAX_LIFETIME")
	cfg.JWT.Secret = v.GetString("JWT_SECRET")
	cfg.JWT.AccessTokenTTL = v.GetDuration("JWT_ACCESS_TOKEN_TTL")

	return cfg, nil
}
