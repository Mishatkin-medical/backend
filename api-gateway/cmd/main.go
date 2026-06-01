package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/golang-jwt/jwt/v5"

	"mishatkin/api-gateway/internal/config"
	zerolog "mishatkin/shared/pkg/logger"
)

type Gateway struct {
	log       *zerolog.Logger
	cfg       *config.Config
	jwtSecret []byte
	services  map[string]string
}

func main() {
	log := zerolog.New("api-gateway")
	log.Info().Msg("Starting API Gateway with Fiber...")

	cfg := config.Load()
	log.SetLevel(cfg.LogLevel)

	g := &Gateway{
		log:       log,
		cfg:       cfg,
		jwtSecret: []byte(cfg.JWTSecret),
		services:  make(map[string]string),
	}

	// Initialize services URLs
	g.services["auth"] = cfg.AuthServiceURL
	g.services["user"] = cfg.UserServiceURL
	g.services["device"] = cfg.DeviceServiceURL
	g.services["glucose"] = cfg.GlucoseServiceURL
	g.services["insulin"] = cfg.InsulinServiceURL
	g.services["delivery"] = cfg.DeliveryServiceURL
	g.services["notification"] = cfg.NotificationServiceURL
	g.services["history"] = cfg.HistoryServiceURL

	// Create Fiber app with optimal settings
	app := fiber.New(fiber.Config{
		AppName:               "mishatkin-api-gateway",
		DisableStartupMessage: false,
		ReadTimeout:           15 * time.Second,
		WriteTimeout:          15 * time.Second,
		IdleTimeout:           60 * time.Second,
		BodyLimit:             4 * 1024 * 1024, // 4MB
	})

	// Global middleware stack
	app.Use(recover.New())
	app.Use(fiberlogger.New(fiberlogger.Config{
		Format: "[${time}] ${status} - ${method} ${path} - ${latency}\n",
	}))

	// Global rate limit by IP
	app.Use(limiter.New(limiter.Config{
		Max:        cfg.RateLimitPerIP,
		Expiration: 1 * time.Minute,
	}))

	// Health endpoints
	app.Get("/api/v1/health/live", g.healthLive)
	app.Get("/api/v1/health/ready", g.healthReady)

	// Auth routes (no auth required) - proxy to auth service
	app.All("/api/v1/auth/*", g.proxyToService("auth"))

	// Protected routes group
	api := app.Group("/api/v1", g.authMiddleware())

	// Rate limit per user for protected routes
	api.Use(limiter.New(limiter.Config{
		Max:        cfg.RateLimitPerUser,
		Expiration: 1 * time.Minute,
	}))

	// Proxy all other requests based on path
	api.All("/*", g.proxyFromPath())

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Info().Msg("Shutting down gateway...")
		if err := app.Shutdown(); err != nil {
			log.Error().Err(err).Msg("Error during shutdown")
		}
	}()

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Info().Str("addr", addr).Msg("HTTP server starting...")
	if err := app.Listen(addr); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("HTTP server error")
	}
}

func (g *Gateway) authMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
return c.Status(fiber.StatusUnauthorized).JSON(map[string]string{
				"error": "missing authorization header",
			})
		}

		// Extract token
		tokenString := authHeader
		if len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = authHeader[7:]
		}

		// Validate JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("invalid signing method")
			}
			return g.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(map[string]string{
				"error": "invalid token",
			})
		}

		// Store user info in locals
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Locals("user_id", claims["sub"])
			c.Locals("claims", claims)
		}

		return c.Next()
	}
}

func (g *Gateway) proxyToService(service string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		serviceURL, ok := g.services[service]
		if !ok {
			return c.Status(fiber.StatusNotFound).JSON(map[string]string{
				"error": "service not found",
			})
		}

		return g.proxyRequest(c, serviceURL)
	}
}

func (g *Gateway) proxyFromPath() fiber.Handler {
	return func(c *fiber.Ctx) error {
		service := g.getServiceFromPath(c.Path())
		serviceURL, ok := g.services[service]
		if !ok {
			return c.Status(fiber.StatusNotFound).JSON(map[string]string{
				"error": "service not found: " + service,
			})
		}

		return g.proxyRequest(c, serviceURL)
	}
}

func (g *Gateway) proxyRequest(c *fiber.Ctx, target string) error {
	// Create request
	reqBody := bytes.NewReader(c.Body())
	req, err := http.NewRequestWithContext(c.Context(), c.Method(), target+c.Path(), reqBody)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(map[string]string{
			"error": "failed to create request",
		})
	}

	// Copy headers
	c.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(map[string]string{
			"error": "failed to proxy request",
		})
	}
	defer resp.Body.Close()

	// Copy response status
	c.Response().SetStatusCode(resp.StatusCode)

	// Copy headers
	for k, v := range resp.Header {
		if len(v) > 0 {
			c.Response().Header.Set(k, v[0])
		}
	}

	// Copy body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]string{
			"error": "failed to read response",
		})
	}

	return c.Send(body)
}

func (g *Gateway) getServiceFromPath(path string) string {
	// /api/v1/service/... -> service
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/"), "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func (g *Gateway) healthLive(c *fiber.Ctx) error {
	return c.JSON(map[string]interface{}{
		"status":  "ok",
		"service": "api-gateway",
		"time":    time.Now().Unix(),
	})
}

func (g *Gateway) healthReady(c *fiber.Ctx) error {
	status := map[string]interface{}{
		"status":   "degraded",
		"services": map[string]string{},
	}

	client := &http.Client{Timeout: 3 * time.Second}
	allUp := true

	for name, url := range g.services {
		resp, err := client.Get(url + "/api/v1/health/live")
		if err != nil || resp == nil || resp.StatusCode != 200 {
			if m, ok := status["services"].(map[string]string); ok {
				m[name] = "down"
			}
			allUp = false
		} else {
			resp.Body.Close()
			if m, ok := status["services"].(map[string]string); ok {
				m[name] = "up"
			}
		}
	}

	if allUp {
		status["status"] = "ok"
	}

	return c.JSON(status)
}