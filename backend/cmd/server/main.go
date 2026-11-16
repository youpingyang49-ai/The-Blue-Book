package main

import (
	"log/slog"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/config"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/database"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/router"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect database", "error", err)
	}

	r := router.New(db)

	slog.Info("server starting", "port", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		slog.Error("server failed", "error", err)
	}
}
