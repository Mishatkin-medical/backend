package redis

import (
	"context"
	"fmt"
	"strconv"

	"mishatkin/auth-service/internal/domain"
	"mishatkin/shared/pkg/redis"
)

// TokenCache implements domain.TokenCache
type TokenCache struct {
	client *redis.Client
}

// NewTokenCache creates a new token cache
func NewTokenCache(client *redis.Client) *TokenCache {
	return &TokenCache{client: client}
}

// Set stores value with expiry
func (c *TokenCache) Set(ctx context.Context, key string, value interface{}, expirySeconds int) error {
	return c.client.SetWithExpiry(ctx, key, value, time.Duration(expirySeconds)*time.Second)
}

// Get retrieves value
func (c *TokenCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.GetString(ctx, key)
}

// Delete removes value
func (c *TokenCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Exists checks if key exists
func (c *TokenCache) Exists(ctx context.Context, key string) (bool, error) {
	return c.client.Exists(ctx, key)
}

// Incr increments counter
func (c *TokenCache) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key)
}

// Expire sets expiry
func (c *TokenCache) Expire(ctx context.Context, key string, seconds int) error {
	return c.client.Expire(ctx, key, time.Duration(seconds)*time.Second)
}

// time package
import "time"

