package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	RedisAddr   string
	JWTSecret   string
	AppEnv      string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "bluebook:bluebook@tcp(127.0.0.1:3306)/bluebook?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr:   getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		JWTSecret:   getEnv("JWT_SECRET", "the-blue-book-dev-secret-change-me"),
		AppEnv:      getEnv("APP_ENV", "development"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
