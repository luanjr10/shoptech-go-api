package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort       string
	DBDSN         string
	JWTSecret     string
	JWTTTL        time.Duration
	AdminEmail    string
	AdminPassword string
}

func Load() Config {
	_ = godotenv.Load()

	ttl, err := time.ParseDuration(getEnv("JWT_TTL", "24h"))
	if err != nil {
		ttl = 24 * time.Hour
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		getEnv("DB_USER", "root"),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_HOST", "127.0.0.1"),
		getEnv("DB_PORT", "3306"),
		getEnv("DB_NAME", "shoptech_go"),
	)

	return Config{
		AppPort:       getEnv("APP_PORT", "8080"),
		DBDSN:         dsn,
		JWTSecret:     getEnv("JWT_SECRET", "change-me-in-production"),
		JWTTTL:        ttl,
		AdminEmail:    getEnv("ADMIN_EMAIL", "admin@shoptech.local"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
