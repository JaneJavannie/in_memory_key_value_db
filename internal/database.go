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
	logger       *slog.Logger
	computeLayer compute.Computer
	engine       engine.Engine
}

func NewDatabase(logger *slog.Logger) *Database {
	return &Database{
		logger:       logger,
		computeLayer: compute.NewComputer(logger),
		engine:       engine.NewInMemoryEngine(logger),
	}
}

func (d *Database) HandleRequest(ctx context.Context, text string) (string, error) {
	query, err := d.computeLayer.Compute(ctx, text)
	if err != nil {
		return "", fmt.Errorf("compute: %w", err)
	}

	d.logger.Info("computed successfully", consts.RequestID, ctx.Value(consts.RequestID).(string), "query", query)

	result, err := d.engine.ProcessCommand(ctx, query)
	if err != nil {
		return "", err
	}

	d.logger.Info("engine: process command success", consts.RequestID, ctx.Value(consts.RequestID).(string), "result", result)

	return result, nil
}
