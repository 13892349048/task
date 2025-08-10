package config

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds application configuration loaded from YAML/env.
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

	// Defaults (dot-keys to align with YAML structure)
	v.SetDefault("environment", "dev")
	v.SetDefault("http.port", "8080")

	v.SetDefault("mysql.max_open_conns", 50)
	v.SetDefault("mysql.max_idle_conns", 25)
	v.SetDefault("mysql.conn_max_lifetime", "30m")

	v.SetDefault("jwt.access_token_ttl", "1h")

	v.SetDefault("redis.db", 0)

	v.SetDefault("cache.ttl", "300s")
	v.SetDefault("cache.null_ttl", "30s")
	v.SetDefault("cache.local_cap", 1000)
	v.SetDefault("cache.jitter_sec", 60)

	// YAML config file
	v.SetConfigType("yaml")
	if path := os.Getenv("TASK_CONFIG"); path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("config")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("/etc/task")
	}
	if err := v.ReadInConfig(); err != nil {
		var nf viper.ConfigFileNotFoundError
		if !errors.As(err, &nf) {
			return nil, err
		}
		// if not found, continue with defaults + env
	}

	// Env overrides (TASK_HTTP_PORT overrides http.port, etc.)
	v.SetEnvPrefix("TASK")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	cfg := &Config{}
	cfg.Environment = v.GetString("environment")
	cfg.HTTPPort = v.GetString("http.port")

	cfg.MySQL.DSN = v.GetString("mysql.dsn")
	cfg.MySQL.MaxOpenConns = v.GetInt("mysql.max_open_conns")
	cfg.MySQL.MaxIdleConns = v.GetInt("mysql.max_idle_conns")
	cfg.MySQL.ConnMaxLifetime = v.GetDuration("mysql.conn_max_lifetime")

	cfg.JWT.Secret = v.GetString("jwt.secret")
	cfg.JWT.AccessTokenTTL = v.GetDuration("jwt.access_token_ttl")

	cfg.Redis.Addr = v.GetString("redis.addr")
	cfg.Redis.Password = v.GetString("redis.password")
	cfg.Redis.DB = v.GetInt("redis.db")

	cfg.Cache.TTL = v.GetDuration("cache.ttl")
	cfg.Cache.NullTTL = v.GetDuration("cache.null_ttl")
	cfg.Cache.LocalCap = v.GetInt("cache.local_cap")
	cfg.Cache.JitterSec = v.GetInt("cache.jitter_sec")

	return cfg, nil
}
