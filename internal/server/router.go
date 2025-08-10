package server

import (
	"net/http"

	"task/internal/handler"
	"task/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type RouterDeps struct {
	AuthHandler *handler.AuthHandler
	TaskHandler *handler.TaskHandler
	Health      *handler.HealthHandler
	JWTSecret   string
}

func NewRouter(deps RouterDeps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Next()
	})
	// Order matters: request id -> logging -> others
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger())

	r.GET("/api/v1/health", deps.Health.Health)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		auth.POST("/login", deps.AuthHandler.Login)

		users := api.Group("/users")
		users.POST("/register", deps.AuthHandler.Register)

		jwtmw := middleware.NewJWT(deps.JWTSecret)
		tasks := api.Group("/tasks")
		tasks.Use(jwtmw.Handle())
		tasks.POST("", deps.TaskHandler.Create)
		tasks.GET("/:task_id", deps.TaskHandler.Get)
	}

	r.NoRoute(func(c *gin.Context) { c.JSON(http.StatusNotFound, gin.H{"error": "not found"}) })
	return r
}
