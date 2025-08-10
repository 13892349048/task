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

	Redis struct {
		Addr     string
		Password string
		DB       int
	}

	Cache struct {
		TTL       time.Duration
		NullTTL   time.Duration
		LocalCap  int
		JitterSec int
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

	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("CACHE_TTL", "300s")
	v.SetDefault("CACHE_NULL_TTL", "30s")
	v.SetDefault("CACHE_LOCAL_CAP", 1000)
	v.SetDefault("CACHE_JITTER_SEC", 60)

	cfg := &Config{}
	cfg.Environment = v.GetString("ENVIRONMENT")
	cfg.HTTPPort = v.GetString("HTTP_PORT")

	cfg.MySQL.DSN = v.GetString("MYSQL_DSN")
	cfg.MySQL.MaxOpenConns = v.GetInt("MYSQL_MAX_OPEN_CONNS")
	cfg.MySQL.MaxIdleConns = v.GetInt("MYSQL_MAX_IDLE_CONNS")
	cfg.MySQL.ConnMaxLifetime = v.GetDuration("MYSQL_CONN_MAX_LIFETIME")

	cfg.JWT.Secret = v.GetString("JWT_SECRET")
	cfg.JWT.AccessTokenTTL = v.GetDuration("JWT_ACCESS_TOKEN_TTL")

	cfg.Redis.Addr = v.GetString("REDIS_ADDR")
	cfg.Redis.Password = v.GetString("REDIS_PASSWORD")
	cfg.Redis.DB = v.GetInt("REDIS_DB")

	cfg.Cache.TTL = v.GetDuration("CACHE_TTL")
	cfg.Cache.NullTTL = v.GetDuration("CACHE_NULL_TTL")
	cfg.Cache.LocalCap = v.GetInt("CACHE_LOCAL_CAP")
	cfg.Cache.JitterSec = v.GetInt("CACHE_JITTER_SEC")

	return cfg, nil
}
