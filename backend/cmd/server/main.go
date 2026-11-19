package main

import (
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/config"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/database"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/model"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/pkg/jwt"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/pkg/storage"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/router"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect database", "error", err)
		os.Exit(1)
	}

	if err := db.AutoMigrate(&model.User{}, &model.Post{}, &model.PostImage{}, &model.Like{}, &model.Favorite{}, &model.Comment{}, &model.Follow{}, &model.Topic{}, &model.PostTopic{}, &model.Notification{}, &model.Message{}, &model.ConversationHide{}); err != nil {
		slog.Error("failed to migrate", "error", err)
		os.Exit(1)
	}

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})

	var st storage.Storage
	if cfg.MinIO.Endpoint != "" {
		st, err = storage.NewMinio(cfg.MinIO.Endpoint, cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, cfg.MinIO.Bucket, cfg.MinIO.UseSSL)
	} else {
		st, err = storage.NewLocal(cfg.UploadDir, cfg.PublicBaseURL)
	}
	if err != nil {
		slog.Error("failed to init storage", "error", err)
		os.Exit(1)
	}

	jwtMgr := jwt.NewManager(cfg.JWTSecret)

	r := router.New(db, rdb, jwtMgr, st, cfg.AppEnv)
	r.Static("/uploads", cfg.UploadDir)

	slog.Info("server starting", "port", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
