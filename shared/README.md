# Shared Library

## Overview

Path: `backend/shared/`

Reusable packages for all microservices. This is the foundation - all 9 services depend on it.

## Packages

### pkg/logger

Structured JSON logging with zerolog.

```go
import "mishatkin/shared/pkg/logger"

log := logger.New("service-name")
log.Info().Str("user_id", userID.String()).Msg("login success")
log.Error().Err(err).Str("endpoint", "/auth/login").Msg("request failed")
```

Output:
```json
{"level":"info","time":"2024-01-01T00:00:00Z","service":"auth-service","user_id":"...","message":"login success"}
```

Features:
- Request ID field (from context)
- Structured fields
- JSON output (machine-readable)
- Log levels: debug, info, warn, error, fatal

### pkg/postgres

Connection pool and helpers.

```go
import "mishatkin/shared/pkg/postgres"

pool, err := postgres.NewPool(ctx, os.Getenv("DB_URL"))
defer pool.Close()

// Run migrations
err = postgres.RunMigrations(ctx, pool, "migrations/001.sql")

// Transaction helper
result, err := postgres.WithTx(ctx, pool, func(tx pgx.Tx) error {
    return repository.InsertUser(ctx, tx, user)
})
```

Features:
- pgxpool connection pool
- Health check
- Migration runner
- Transaction wrapper
- Query timeout

### pkg/redis

Redis client with Sentinel support.

```go
import "mishatkin/shared/pkg/redis"

client, err := redis.NewClient(os.Getenv("REDIS_URL"))
defer client.Close()

// Distributed lock
lock, err := client.AcquireLock(ctx, "delivery:nonce:123", 5*time.Minute)
if err != nil {
    return ErrNonceAlreadyUsed
}
defer lock.Release()
```

Features:
- Sentinel connection
- Distributed locks (Redlock)
- Rate limiting helpers
- Cache with TTL

### pkg/jwt

JWT token creation and validation.

```go
import "mishatkin/shared/pkg/jwt"

tokens, err := jwt.IssueTokens(ctx, jwt.IssueInput{
    UserID: user.ID,
    Email:  user.Email,
    Secret: os.Getenv("JWT_SECRET"),
})
// tokens.AccessToken, tokens.RefreshToken

claims, err := jwt.ValidateAccess(ctx, token, secret)
```

Features:
- Access token (15 min)
- Refresh token (30 days)
- Claims extraction
- Token validation

### pkg/errors

Standard error codes and HTTP mapping.

```go
import "mishatkin/shared/pkg/errors"

// Domain errors
var (
    ErrUserNotFound     = errors.New("USER_NOT_FOUND", 404)
    ErrInvalidPassword   = errors.New("INVALID_PASSWORD", 401)
    ErrTokenExpired      = errors.New("TOKEN_EXPIRED", 401)
)

// HTTP handler
if errors.Is(err, ErrUserNotFound) {
    respondError(w, err)
}
```

Features:
- Error codes
- HTTP status mapping
- Error wrapping
- Client error formatting

### pkg/validator

Struct validation.

```go
import "mishatkin/shared/pkg/validator"

type RegisterRequest struct {
    Email    string `validate:"required,email"`
    Phone    string `validate:"omitempty,e164"`
    Password string `validate:"required,min=8"`
}

err := validator.Validate(req)
```

Uses go-playground/validator v10.

## Proto Files

All .proto definitions for gRPC inter-service communication.

```
proto/
  auth.proto         -- AuthService
  user.proto         -- UserService
  device.proto       -- DeviceService
  glucose.proto      -- GlucoseService
  insulin.proto      -- InsulinService
  delivery.proto     -- DeliveryService
  notification.proto -- NotificationService
  history.proto      -- HistoryService
```

Generate Go code:
```bash
cd backend/shared
make generate
# Runs: protoc --go_out=. --go-grpc_out=. proto/*.proto
```

## Makefile Targets

```bash
make generate    # Generate .pb.go from .proto
make tidy        # go mod tidy
make lint        # golangci-lint run
make test        # go test ./...
```

## Import in services

```go
require mishatkin/shared v0.0.0
```

Replace version with git tag as releases are cut.

## File Structure

```
shared/
  pkg/
    logger/
      logger.go
      fields.go
    postgres/
      pool.go
      migrations.go
      tx.go
    redis/
      client.go
      distributed_lock.go
    jwt/
      jwt.go
      claims.go
    errors/
      errors.go
      codes.go
    validator/
      validator.go
    tracing/
      otel.go
  proto/
    auth.proto
    user.proto
    device.proto
    glucose.proto
    insulin.proto
    delivery.proto
    notification.proto
    history.proto
  Makefile
  go.mod
  go.sum
  README.md
```

## Version

Current: v0.1.0 (Phase 1 development)

## Related

- `docs/services/shared.md`
- `docs/architecture/ADR-001-architecture-overview.md`

