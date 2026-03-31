package observability

import (
	"log/slog"
	"os"
	"strings"
)

const (
	LogLevelDebug = "DEBUG"
	LogLevelWarn  = "WARN"
	LogLevelInfo  = "INFO"
	LogLevelError = "ERROR"

	LogFormatJSON = "JSON"
)

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

var Logger *slog.Logger = slog.New(getHandler())
