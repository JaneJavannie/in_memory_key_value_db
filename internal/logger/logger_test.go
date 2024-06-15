package logger

import (
	"log/slog"
	"testing"

	"github.com/JaneJavannie/in_memory_key_value_db/internal/configs"
)

func TestNewLogger(t *testing.T) {
	config := configs.Logger{
		Level:    "info",
		IsPretty: true,
	}

	_, err := NewLogger(config)
	if err != nil {
		t.Errorf("Error creating logger: %v", err)
	}
}

func TestParseLevel(t *testing.T) {
	level, err := parseLevel("debug")
	if err != nil {
		t.Errorf("Error parsing level: %v", err)
	}
	if level != slog.LevelDebug {
		t.Errorf("Expected level Debug, got %v", level)
	}
}
