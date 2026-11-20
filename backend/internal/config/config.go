package config

import (
	"os"
	"time"
)

type Config struct {
	Port            string
	DatabaseURL     string
	RedisAddr       string
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	AppEnv          string
	UploadDir       string
	PublicBaseURL   string
	RabbitMQURL     string
	MinIO           MinIOConfig
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", "bluebook:bluebook@tcp(127.0.0.1:3306)/bluebook?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr:       getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		JWTSecret:       getEnv("JWT_SECRET", "the-blue-book-dev-secret-change-me"),
		AccessTokenTTL:  getDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL: getDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour),
		AppEnv:          getEnv("APP_ENV", "development"),
		UploadDir:       getEnv("UPLOAD_DIR", "./uploads"),
		PublicBaseURL:   getEnv("PUBLIC_BASE_URL", "http://127.0.0.1:8080"),
		RabbitMQURL:     getEnv("RABBITMQ_URL", "amqp://guest:guest@127.0.0.1:5672/"),
		MinIO: MinIOConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", ""),
			AccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			Bucket:    getEnv("MINIO_BUCKET", "bluebook"),
			UseSSL:    getEnv("MINIO_USE_SSL", "false") == "true",
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}
