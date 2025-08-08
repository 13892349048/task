package main

import (
	"net/http"
	"time"

	"task/internal/config"
	"task/internal/handler"
	"task/internal/repository"
	"task/internal/server"
	"task/internal/service"
	"task/pkg/logger"

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

	authSvc := service.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.AccessTokenTTL)
	taskSvc := service.NewTaskService(taskRepo)

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
