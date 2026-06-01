package postgres

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool wraps pgxpool.Pool with InsulinPro-specific config
type Pool struct {
	*pgxpool.Pool
}

// Config holds connection configuration
type Config struct {
	URL        string
	MaxConns   int32
	MinConns   int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
	HealthCheckPeriod time.Duration
}

// DefaultConfig returns config from environment
func DefaultConfig() *Config {
	return &Config{
		URL:               os.Getenv("DB_URL"),
		MaxConns:          20,
		MinConns:          5,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: time.Minute,
	}
}

// NewPool creates a new connection pool
func NewPool(ctx context.Context, cfg *Config) (*Pool, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("DB_URL is required")
	}

	config, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DB_URL: %w", err)
	}

	config.MaxConns = cfg.MaxConns
	config.MinConns = cfg.MinConns
	config.MaxConnLifetime = cfg.MaxConnLifetime
	config.MaxConnIdleTime = cfg.MaxConnIdleTime
	config.HealthCheckPeriod = cfg.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Pool{Pool: pool}, nil
}

// HealthCheck checks if database is reachable
func (p *Pool) HealthCheck(ctx context.Context) error {
	return p.Ping(ctx)
}

// Close closes the pool
func (p *Pool) Close() {
	p.Pool.Close()
}
