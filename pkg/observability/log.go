// Package observability sets up structured logging (`log/slog`) with a
// JSON handler at the configured level and ties the logger to the
// service name. OTEL traces and Prometheus metrics are exposed by
// sibling files.
package observability

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

// LogConfig is the input to NewLogger.
type LogConfig struct {
	// Level is "debug" | "info" | "warn" | "error" (case-insensitive).
	Level string
	// Service is the service name attached to every record as
	// `service.name`.
	Service string
	// Environment is "dev" | "staging" | "prod".
	Environment string
	// Writer overrides the destination (default os.Stdout).
	Writer io.Writer
}

// NewLogger returns a *slog.Logger producing JSON records with
// `service.name`, `service.environment`, and `service.commit` baked
// into every entry.
func NewLogger(cfg LogConfig) *slog.Logger {
	w := cfg.Writer
	if w == nil {
		w = os.Stdout
	}
	level := parseLevel(cfg.Level)
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			// Compact "time" key to "ts" and "msg" stays as-is — the most
			// common log shipper schema.
			if a.Key == slog.TimeKey {
				return slog.Attr{Key: "ts", Value: a.Value}
			}
			return a
		},
	})
	logger := slog.New(handler).With(
		slog.String("service.name", cfg.Service),
		slog.String("service.environment", cfg.Environment),
	)
	return logger
}

// LoggerFromContext returns the logger stored in `ctx` via
// [WithLogger], or [slog.Default] if absent.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// WithLogger returns a copy of ctx that carries `logger`.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

type loggerKey struct{}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
