package compute

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/JaneJavannie/in_memory_key_value_db/internal/consts"
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
	parser   *Parser
	analyzer *Analyzer
	logger   *slog.Logger
}

func NewComputer(logger *slog.Logger) Computer {
	p := newParser(logger)
	a := newAnalyzer(logger)

	return Computer{
		parser:   &p,
		analyzer: &a,
		logger:   logger,
	}
}

func (c *Computer) Compute(ctx context.Context, text string) (Query, error) {
	c.logger.Debug("parsing text", consts.RequestID, ctx.Value(consts.RequestID).(string), "text", text)

	result, err := c.parser.parse(text)
	if err != nil {
		return Query{}, fmt.Errorf("parse: %w", err)
	}

	c.logger.Debug("parse success", consts.RequestID, ctx.Value(consts.RequestID).(string))

	c.logger.Debug("analyzing parse result", consts.RequestID, ctx.Value(consts.RequestID).(string), "result", result)

	query, err := c.analyzer.analyzeQuery(ctx, result)
	if err != nil {
		return Query{}, fmt.Errorf("analyze: %w", err)
	}

	c.logger.Debug("analyze success", consts.RequestID, ctx.Value(consts.RequestID).(string))

	return query, nil
}
