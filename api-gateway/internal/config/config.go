package config

import "os"
import "strconv"

type Config struct {
	Port            int
	LogLevel        string
	JWTSecret       string
	RateLimitPerUser int
	RateLimitPerIP   int
	AuthServiceURL       string
	UserServiceURL       string
	DeviceServiceURL     string
	GlucoseServiceURL     string
	InsulinServiceURL     string
	DeliveryServiceURL    string
	NotificationServiceURL string
	HistoryServiceURL     string
	RedisURL             string
}

func Load() *Config {
	port := 8080
	if p := os.Getenv("GATEWAY_PORT"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			port = parsed
		}
	}

	return &Config{
		Port:             port,
		LogLevel:         getEnv("GATEWAY_LOG_LEVEL", "info"),
		JWTSecret:        getEnv("GATEWAY_JWT_SECRET", ""),
		RateLimitPerUser: intEnv("GATEWAY_MISHATKIN_RATE_LIMIT_PER_USER", 100),
		RateLimitPerIP:   intEnv("GATEWAY_MISHATKIN_RATE_LIMIT_PER_IP", 500),
		AuthServiceURL:        getEnv("MISHATKIN_AUTH_SERVICE_URL", "http://auth-service:8081"),
		UserServiceURL:        getEnv("MISHATKIN_USER_SERVICE_URL", "http://user-service:8082"),
		DeviceServiceURL:      getEnv("MISHATKIN_DEVICE_SERVICE_URL", "http://device-service:8083"),
		GlucoseServiceURL:      getEnv("MISHATKIN_GLUCOSE_SERVICE_URL", "http://glucose-service:8084"),
		InsulinServiceURL:      getEnv("MISHATKIN_INSULIN_SERVICE_URL", "http://insulin-service:8085"),
		DeliveryServiceURL:     getEnv("MISHATKIN_DELIVERY_SERVICE_URL", "http://delivery-service:8086"),
		NotificationServiceURL: getEnv("MISHATKIN_NOTIFICATION_SERVICE_URL", "http://notification-service:8087"),
		HistoryServiceURL:      getEnv("MISHATKIN_HISTORY_SERVICE_URL", "http://history-service:8088"),
		RedisURL:              getEnv("MISHATKIN_REDIS_URL", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func intEnv(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return defaultValue
}

