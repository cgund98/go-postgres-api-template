package observability

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

const (
	LogLevelDebug = "DEBUG"
	LogLevelWarn  = "WARN"
	LogLevelInfo  = "INFO"
	LogLevelError = "ERROR"

	LogFormatJSON = "JSON"
)

var loggerCtxKey = &struct{ name string }{"structured_logger"}

func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerCtxKey).(*slog.Logger); ok {
		return logger
	}
	return nil
}

func SetLoggerOnContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey, logger)
}

func getLevel() slog.Level {
	levelStr := os.Getenv("LOG_LEVEL")

	if strings.ToUpper(levelStr) == LogLevelDebug {
		return slog.LevelDebug
	}
	if strings.ToUpper(levelStr) == LogLevelWarn {
		return slog.LevelWarn
	}
	if strings.ToUpper(levelStr) == LogLevelInfo {
		return slog.LevelInfo
	}
	if strings.ToUpper(levelStr) == LogLevelError {
		return slog.LevelError
	}
	return slog.LevelInfo
}

func getHandler() slog.Handler {
	logFormat := strings.ToUpper(os.Getenv("LOG_FORMAT"))
	if logFormat == LogFormatJSON {
		return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: getLevel(),
		})
	}

	return slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: getLevel(),
	})
}

func NewBootstrapLogger() *slog.Logger {
	return slog.New(getHandler())
}

func NewLogger(serviceName string) *slog.Logger {
	return otelslog.NewLogger(serviceName, otelslog.WithSource(true))
}
