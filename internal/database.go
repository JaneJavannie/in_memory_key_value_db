package internal

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/JaneJavannie/in_memory_key_value_db/internal/compute"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/consts"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/storage/engine"
)

type computeLayer interface {
	Compute(ctx context.Context, text string) (compute.Query, error)
}

type engineLayer interface {
	ProcessCommand(ctx context.Context, query compute.Query) (string, error)
}

type databaseLayer interface {
	HandleRequest(ctx context.Context, text string) (string, error)
}

type Database struct {
	engine       *engine.Engine
	computeLayer compute.Computer
	logger       *slog.Logger
}

func NewDatabase(engine *engine.Engine, logger *slog.Logger) (*Database, error) {
	return &Database{
		engine:       engine,
		computeLayer: compute.NewComputer(logger),
		logger:       logger,
	}, nil
}

func (d *Database) HandleRequest(ctx context.Context, text string) (string, error) {
	query, err := d.computeLayer.Compute(ctx, text)
	if err != nil {
		return "", fmt.Errorf("compute: %w", err)
	}

	d.logger.Info("computed successfully", consts.RequestID, ctx.Value(consts.RequestID).(string), "query", query)

	result, err := d.engine.ProcessCommand(ctx, query)
	if err != nil {
		return "", fmt.Errorf("process command: %w", err)
	}

	d.logger.Info("engine: process command success", consts.RequestID, ctx.Value(consts.RequestID).(string), "result", result)

	return result, nil
}
