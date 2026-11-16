package router

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/handler"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/middleware"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/pkg/jwt"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/repository"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/service"
)

func New(db *gorm.DB, rdb *redis.Client, jwtMgr *jwt.Manager, appEnv string) *gin.Engine {
	r := gin.Default()

	health := handler.NewHealthHandler(db)

	userRepo := repository.NewUserRepo(db)
	authSvc := service.NewAuthService(userRepo, rdb, jwtMgr)
	authHandler := handler.NewAuthHandler(authSvc, appEnv)
	userHandler := handler.NewUserHandler(userRepo)

	api := r.Group("/api/v1")
	{
		api.GET("/health", health.Check)

		auth := api.Group("/auth")
		{
			auth.POST("/send-code", authHandler.SendCode)
			auth.POST("/login", authHandler.Login)
		}

		protected := api.Group("")
		protected.Use(middleware.Auth(jwtMgr))
		{
			protected.GET("/users/me", userHandler.Me)
		}
	}

	return r
}
