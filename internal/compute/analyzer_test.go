package compute

import (
	"context"
	"testing"

	"in_memory_key_value_db/internal/consts"
)

func TestAnalyzeQuery(t *testing.T) {
	a := &Analyzer{}

	t.Run("TestAnalyzeQuery_ValidSetCommand", func(t *testing.T) {
		parsed := []string{"SET", "key", "value"}
		q, err := a.analyzeQuery(context.Background(), parsed)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if q.Command != "SET" {
			t.Errorf("Expected command to be 'SET', got: %s", q.Command)
		}
		if len(q.Arguments) != 2 {
			t.Errorf("Expected 2 arguments, got: %d", len(q.Arguments))
		}
	})

	t.Run("TestAnalyzeQuery_InvalidGetCommand", func(t *testing.T) {
		parsed := []string{"GET"}
		_, err := a.analyzeQuery(context.Background(), parsed)

		if err == nil {
			t.Errorf("Expected error: %v, got: %v", consts.ErrInvalidGetQueryArgs, err)
		}
	})

	t.Run("TestAnalyzeQuery_EmptyParsedArray", func(t *testing.T) {
		parsed := []string{}
		_, err := a.analyzeQuery(context.Background(), parsed)

		if err == nil {
			t.Errorf("Expected error: 'empty parsed query args array', got: %v", err)
		}
	})
}

func TestValidate(t *testing.T) {
	a := &Analyzer{}

	t.Run("TestValidate_EmptyParsedArray", func(t *testing.T) {
		parsed := []string{}
		err := a.validate(context.Background(), parsed)

		if err == nil || err.Error() != "empty parsed query args array" {
			t.Errorf("Expected error: 'empty parsed query args array', got: %v", err)
		}
	})

	t.Run("TestValidate_SetCommandInvalidArgs", func(t *testing.T) {
		parsed := []string{"SET", "key"}
		err := a.validate(context.Background(), parsed)

		if err == nil {
			t.Errorf("Expected error: %v, got: %v", consts.ErrInvalidSetQueryArgs, err)
		}
	})

	t.Run("TestValidate_GetCommandInvalidArgs", func(t *testing.T) {
		parsed := []string{"GET"}
		err := a.validate(context.Background(), parsed)

		if err == nil {
			t.Errorf("Expected error: %v, got: %v", consts.ErrInvalidGetQueryArgs, err)
		}
	})

	t.Run("TestValidate_DelCommandInvalidArgs", func(t *testing.T) {
		parsed := []string{"DEL"}
		err := a.validate(context.Background(), parsed)

		if err == nil {
			t.Errorf("Expected error: %v, got: %v", consts.ErrInvalidDelQueryArgs, err)
		}
	})

	t.Run("TestValidate_UnknownCommand", func(t *testing.T) {
		parsed := []string{"UNKNOWN"}
		err := a.validate(context.Background(), parsed)

		if err == nil || err.Error() != "unknown command: UNKNOWN" {
			t.Errorf("Expected error: 'unknown command: UNKNOWN', got: %v", err)
		}
	})
}
