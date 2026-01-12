package observability

import (
	"log/slog"
	"os"
	"strings"
)

func getLevel() slog.Level {
	levelStr := os.Getenv("LOG_LEVEL")

	if strings.ToUpper(levelStr) == "DEBUG" {
		return slog.LevelDebug
	}
	if strings.ToUpper(levelStr) == "WARN" {
		return slog.LevelWarn
	}
	return slog.LevelInfo
}

func getHandler() slog.Handler {

	logForamt := os.Getenv("LOG_FORMAT")
	if logForamt == "json" {
		return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: getLevel(),
		})
	}
	return slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: getLevel(),
	})
}

var Logger *slog.Logger = slog.New(getHandler())
