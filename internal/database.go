package internal

import (
	"context"

	"in_memory_key_value_db/internal/compute"
	"in_memory_key_value_db/internal/storage/engine"
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
	computeLayer compute.Computer
	engine       engine.Engine
}

func NewDatabase() *Database {
	return &Database{
		computeLayer: compute.NewComputer(),
		engine:       engine.NewInMemoryEngine(),
	}
}

func (d *Database) HandleRequest(ctx context.Context, text string) (string, error) {
	//c := compute.NewComputer()
	query, err := d.computeLayer.Compute(ctx, text)
	if err != nil {
		return "", err
	}

	//e := engine.NewInMemoryEngine()

	result, err := d.engine.ProcessCommand(ctx, query)
	if err != nil {
		return "", err
	}

	return result, nil
}
