package compute

import (
	"context"
	"fmt"
	"strings"

	"in_memory_key_value_db/internal/consts"
)

type Query struct {
	Command   string
	Arguments []string
}

// Analyzer компонент внутри слоя, отвечающий за анализ запроса
type Analyzer struct{}

func newAnalyzer() Analyzer { return Analyzer{} }

// AnalyzeQuery из массива строк делает команду и аргументы
func (a *Analyzer) analyzeQuery(ctx context.Context, parsed []string) (Query, error) {
	err := a.validate(ctx, parsed)
	if err != nil {
		return Query{}, fmt.Errorf("validate: %w", err)
	}

	q := Query{
		Command:   strings.ToUpper(parsed[0]),
		Arguments: parsed[1:],
	}

	return q, nil
}

func (a *Analyzer) validate(ctx context.Context, parsed []string) error {
	if len(parsed) == 0 {
		return fmt.Errorf("empty parsed query args array")
	}

	command := strings.ToUpper(parsed[0])

	switch command {
	case consts.CommandSet:
		if len(parsed) != 3 {
			return consts.ErrInvalidSetQueryArgs
		}
	case consts.CommandGet:
		if len(parsed) != 2 {
			return consts.ErrInvalidGetQueryArgs
		}
	case consts.CommandDel:
		if len(parsed) != 2 {
			return consts.ErrInvalidDelQueryArgs
		}
	default:
		return fmt.Errorf("unknown command: %s", command)
	}

	return nil
}
