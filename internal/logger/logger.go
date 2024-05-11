package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"in_memory_key_value_db/internal/configs"

	"github.com/lmittmann/tint"
)

func NewLogger(config configs.Logger) (*slog.Logger, error) {
	lvl, err := parseLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("error parsing log level: %w", err)
	}

	var h slog.Handler

	h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl, AddSource: false})

	if config.IsPretty {
		h = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      lvl,
			TimeFormat: time.TimeOnly,
			AddSource:  false,
		})
	}

	return slog.New(h), nil
}

func parseLevel(level string) (slog.Level, error) {
	switch strings.ToUpper(level) {
	case slog.LevelDebug.String():
		return slog.LevelDebug, nil
	case slog.LevelInfo.String():
		return slog.LevelInfo, nil
	case slog.LevelWarn.String():
		return slog.LevelWarn, nil
	case slog.LevelError.String():
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid log level: %s", level)
	}
}
