package compute

import (
	"context"
	"fmt"
	"log/slog"

	"in_memory_key_value_db/internal/consts"
)

// compute - слой, отвечающий за обработку запроса

type parser interface {
	parse(string) ([]string, error)
}

type analyzer interface {
	analyzeQuery(ctx context.Context, parsed []string) (Query, error)
	validate(ctx context.Context, parsed []string) error
}

type Computer struct {
	logger *slog.Logger
}

func NewComputer(logger *slog.Logger) Computer {
	return Computer{
		logger: logger,
	}
}

func (c *Computer) Compute(ctx context.Context, text string) (Query, error) {
	p := newParser(c.logger)

	c.logger.Debug("parsing text", consts.RequestID, ctx.Value(consts.RequestID).(string), "text", text)

	result, err := p.parse(text)
	if err != nil {
		return Query{}, fmt.Errorf("parse: %w", err)
	}

	c.logger.Debug("parse success", consts.RequestID, ctx.Value(consts.RequestID).(string))

	a := newAnalyzer(c.logger)

	c.logger.Debug("analyzing parse result", consts.RequestID, ctx.Value(consts.RequestID).(string), "result", result)

	query, err := a.analyzeQuery(ctx, result)
	if err != nil {
		return Query{}, fmt.Errorf("analyze: %w", err)
	}

	c.logger.Debug("analyze success", consts.RequestID, ctx.Value(consts.RequestID).(string))

	return query, nil
}
