package platform

import (
	"github.com/bmbl-bumble2/recs-votes-storage/config"
	"log/slog"
	"os"
	"strings"
)

const (
	LogLevelDebug = "DEBUG"
	LogLevelWarn  = "WARN"
	LogLevelError = "ERROR"
)

//go:generate mockgen -destination=../../testlib/mocks/logger_mock.go -package=mocks github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform Logger
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

func NewLogger(config config.Config) Logger {
	level := strings.ToUpper(config.LogLevel)

	var slogLevel slog.Level
	switch level {
	case LogLevelDebug:
		slogLevel = slog.LevelDebug
	case LogLevelWarn:
		slogLevel = slog.LevelWarn
	case LogLevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLevel,
	})

	logger := slog.New(handler)
	logger.Info("Logger initialized", "level", slogLevel.String())

	return logger
}
