package engine

import (
	"context"
	"fmt"
	"log/slog"

	"in_memory_key_value_db/internal/compute"
	"in_memory_key_value_db/internal/consts"
)

type Engine struct {
	storage *InMemoryStorage
}

func NewInMemoryEngine() Engine {
	return Engine{
		storage: InitMemoryStorage(),
	}
}

func (e *Engine) ProcessCommand(ctx context.Context, query compute.Query) (string, error) {
	queryResult := ""
	var err error

	slog.Debug("processing command", consts.RequestID, ctx.Value(consts.RequestID).(string), "command", query.Command)

	switch query.Command {

	case consts.CommandSet:
		err = e.processSet(ctx, query)
		if err != nil {
			return "", fmt.Errorf("error processing set: %w", err)
		}

	case consts.CommandGet:
		queryResult, err = e.processGet(ctx, query)
		if err != nil {
			return "", fmt.Errorf("error processing get: %w", err)
		}

	case consts.CommandDel:
		err = e.processDel(ctx, query)
		if err != nil {
			return "", fmt.Errorf("error processing del: %w", err)
		}
	}

	return queryResult, nil
}

func (e *Engine) processSet(ctx context.Context, query compute.Query) error {
	e.storage.Set(query.Arguments[0], query.Arguments[1])

	return nil
}

func (e *Engine) processGet(ctx context.Context, query compute.Query) (string, error) {
	val, ok := e.storage.Get(query.Arguments[0])
	if !ok {
		return "", fmt.Errorf("error getting key: key not found")
	}

	return val, nil
}

func (e *Engine) processDel(ctx context.Context, query compute.Query) error {
	e.storage.Del(query.Arguments[0])

	return nil
}
