package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port       int
	LogLevel   string
	DatabaseURL string
	RedisURL    string
	JWTSecret   string
	AccessTTL   string
	RefreshTTL  string
	BcryptCost  int
}

func Load() *Config {
	port := 8081
	if p := os.Getenv("MISHATKIN_AUTH_PORT"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			port = parsed
		}
	}

	accessTTL := "15m"
	if t := os.Getenv("MISHATKIN_JWT_ACCESS_TTL"); t != "" {
		accessTTL = t
	}

	refreshTTL := "720h"
	if t := os.Getenv("MISHATKIN_REFRESH_TTL"); t != "" {
		refreshTTL = t
	}

	bcryptCost := 12
	if c := os.Getenv("MISHATKIN_BCRYPT_COST"); c != "" {
		if parsed, err := strconv.Atoi(c); err == nil {
			bcryptCost = parsed
		}
	}

	return &Config{
		Port:       port,
		LogLevel:   getEnv("MISHATKIN_LOG_LEVEL", "info"),
		DatabaseURL: getEnv("MISHATKIN_DB_URL", "postgres://mishatkin:password@localhost:5432/mishatkin?sslmode=disable"),
		RedisURL:    getEnv("MISHATKIN_REDIS_URL", ""),
		JWTSecret:   getEnv("MISHATKIN_JWT_SECRET", "change-me-in-prod-32-bytes!!"),
		AccessTTL:   accessTTL,
		RefreshTTL:  refreshTTL,
		BcryptCost:  bcryptCost,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return time.Minute * 15
	}
	return d
}


