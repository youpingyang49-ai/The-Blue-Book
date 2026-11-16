package config

import "os"

type Config struct {
	Port          string
	DatabaseURL   string
	RedisAddr     string
	JWTSecret     string
	AppEnv        string
	UploadDir     string
	PublicBaseURL string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "bluebook:bluebook@tcp(127.0.0.1:3306)/bluebook?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr:     getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		JWTSecret:     getEnv("JWT_SECRET", "the-blue-book-dev-secret-change-me"),
		AppEnv:        getEnv("APP_ENV", "development"),
		UploadDir:     getEnv("UPLOAD_DIR", "./uploads"),
		PublicBaseURL: getEnv("PUBLIC_BASE_URL", "http://127.0.0.1:8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
