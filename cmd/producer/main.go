package main

import (
	"net/http"
	"time"

	"task/internal/cache"
	"task/internal/config"
	"task/internal/handler"
	"task/internal/repository"
	"task/internal/server"
	"task/internal/service"
	"task/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log, err := logger.Init(cfg.Environment)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	db, err := repository.NewDB(cfg.MySQL.DSN, cfg.MySQL.MaxOpenConns, cfg.MySQL.MaxIdleConns, cfg.MySQL.ConnMaxLifetime)
	if err != nil {
		log.Fatal("failed to connect db", zap.Error(err))
	}

	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)

	// Cache wiring: prefer local first, then redis
	var store cache.Store
	var local *cache.LocalStore
	if cfg.Cache.LocalCap > 0 {
		local = cache.NewLocalStore(cfg.Cache.LocalCap, cfg.Cache.TTL)
		store = local
	}
	if cfg.Redis.Addr != "" {
		rc := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr, Password: cfg.Redis.Password, DB: cfg.Redis.DB})
		redisStore := cache.NewRedisStore(rc, cfg.Cache.JitterSec)
		store = &cache.MultiStore{Primary: redisStore, Secondary: local}
	}

	var taskSvc *service.TaskService
	if store != nil {
		taskSvc = service.NewTaskServiceWithCache(taskRepo, store, int64(cfg.Cache.TTL.Seconds()), int64(cfg.Cache.NullTTL.Seconds()))
	} else {
		taskSvc = service.NewTaskService(taskRepo)
	}
	authSvc := service.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.AccessTokenTTL)

	authHandler := handler.NewAuthHandler(authSvc)
	taskHandler := handler.NewTaskHandler(taskSvc)
	healthHandler := handler.NewHealthHandler()

	r := server.NewRouter(server.RouterDeps{
		AuthHandler: authHandler,
		TaskHandler: taskHandler,
		Health:      healthHandler,
		JWTSecret:   cfg.JWT.Secret,
	})

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Info("server starting", zap.String("port", cfg.HTTPPort))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("server error", zap.Error(err))
	}
}
