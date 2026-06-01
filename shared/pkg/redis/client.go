package redis

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps redis.Client with InsulinPro-specific helpers
type Client struct {
	*redis.Client
}

// Config holds Redis configuration
type Config struct {
	URL      string
	Password string
	DB       int
}

// DefaultConfig returns config from environment
func DefaultConfig() *Config {
	return &Config{
		URL:      os.Getenv("REDIS_URL"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	}
}

// NewClient creates a new Redis client
func NewClient(cfg *Config) (*Client, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("REDIS_URL is required")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.URL,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{Client: client}, nil
}

// Close closes the client
func (c *Client) Close() error {
	return c.Client.Close()
}

// SetWithExpiry sets a key with expiry
func (c *Client) SetWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	return c.Set(ctx, key, value, expiry).Err()
}

// GetString gets a string value
func (c *Client) GetString(ctx context.Context, key string) (string, error) {
	return c.Get(ctx, key).Result()
}

// Exists checks if key exists
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.Exists(ctx, key).Result()
	return result > 0, err
}

// Incr increments a counter
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.Incr(ctx, key).Result()
}

// SetNX sets value only if not exists (for distributed locks)
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiry time.Duration) (bool, error) {
	return c.SetNX(ctx, key, value, expiry).Result()
}

// AcquireLock acquires a distributed lock
func (c *Client) AcquireLock(ctx context.Context, lockKey string, ttl time.Duration) (*Lock, error) {
	lock := &Lock{
		client: c,
		key:    lockKey,
		ttl:    ttl,
	}
	
	// Try to acquire lock
	acquired, err := c.SetNX(ctx, lockKey, "locked", ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}
	
	if !acquired {
		return nil, fmt.Errorf("lock already held")
	}
	
	return lock, nil
}

// Lock represents a distributed lock
type Lock struct {
	client *Client
	key    string
	ttl    time.Duration
}

// Release releases the lock
func (l *Lock) Release(ctx context.Context) error {
	return l.client.Del(ctx, l.key).Err()
}

// Extend extends the lock TTL
func (l *Lock) Extend(ctx context.Context, ttl time.Duration) error {
	return l.client.Expire(ctx, l.key, ttl).Err()
}

// RateLimitKey generates rate limit key
func RateLimitKey(prefix, identifier string, window time.Duration) string {
	return fmt.Sprintf("ratelimit:%s:%s:%dmin", prefix, identifier, int(window.Minutes()))
}
