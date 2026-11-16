package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/handler"
)

func New(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	health := handler.NewHealthHandler(db)

	api := r.Group("/api/v1")
	{
		api.GET("/health", health.Check)
	}

	return r
}
