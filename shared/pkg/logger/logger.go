package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	LoggerKey    contextKey = "logger"
)

// Logger wraps zerolog.Logger with service name
type Logger struct {
	zl     zerolog.Logger
	svcName string
}

// New creates a new logger for a service
func New(serviceName string) *Logger {
	// Set default output to stdout with JSON format
	zerolog.TimeFieldFormat = time.RFC3339Nano
	
	l := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger()

	return &Logger{
		zl:     l,
		svcName: serviceName,
	}
}

// NewWithWriter creates logger with custom writer (for testing)
func NewWithWriter(w io.Writer, serviceName string) *Logger {
	l := zerolog.New(w).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger()

	return &Logger{
		zl:     l,
		svcName: serviceName,
	}
}

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		requestID = uuid.New().String()
	}
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// FromContext returns logger from context
func FromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(LoggerKey).(*Logger); ok {
		return l
	}
	return New("unknown")
}

// WithContext returns logger with context fields
func (l *Logger) WithContext(ctx context.Context) *Logger {
	requestID := GetRequestID(ctx)
	if requestID != "" {
		l.zl = l.zl.With().Str("request_id", requestID).Logger()
	}
	return l
}

// Debug logs debug level
func (l *Logger) Debug() *zerolog.Event {
	return l.zl.Debug()
}

// Info logs info level
func (l *Logger) Info() *zerolog.Event {
	return l.zl.Info()
}

// Warn logs warning level
func (l *Logger) Warn() *zerolog.Event {
	return l.zl.Warn()
}

// Error logs error level
func (l *Logger) Error() *zerolog.Event {
	return l.zl.Error()
}

// Fatal logs fatal level and exits
func (l *Logger) Fatal() *zerolog.Event {
	return l.zl.Fatal()
}

// SetLevel sets minimum log level
func (l *Logger) SetLevel(level string) {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		l.zl.Warn().Str("level", level).Msg("invalid log level, using info")
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)
}

// LogRequest logs HTTP request
func (l *Logger) LogRequest(ctx context.Context, method, path string, status int, duration time.Duration) {
	l.WithContext(ctx).Info().
		Str("method", method).
		Str("path", path).
		Int("status", status).
		Dur("duration", duration).
		Msg("request")
}

// LogError logs error with context
func (l *Logger) LogError(ctx context.Context, err error, msg string) {
	l.WithContext(ctx).Error().
		Err(err).
		Msg(msg)
}
