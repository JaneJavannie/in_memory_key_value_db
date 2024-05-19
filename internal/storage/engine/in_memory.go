package engine

import (
	"context"
	"log/slog"

	"github.com/JaneJavannie/in_memory_key_value_db/internal/compute"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/consts"
)

type Engine struct {
	logger  *slog.Logger
	storage *InMemoryStorage
}

func NewInMemoryEngine(logger *slog.Logger) Engine {
	return Engine{
		logger:  logger,
		storage: InitMemoryStorage(),
	}
}

func (e *Engine) ProcessCommand(ctx context.Context, query compute.Query) string {
	queryResult := ""

	slog.Debug("processing command", consts.RequestID, ctx.Value(consts.RequestID).(string), "command", query.Command)

	switch query.Command {
	case consts.CommandSet:
		e.processSet(ctx, query)

	case consts.CommandGet:
		queryResult = e.processGet(ctx, query)

	case consts.CommandDel:
		e.processDel(ctx, query)
	}

	return queryResult
}

func (e *Engine) processSet(ctx context.Context, query compute.Query) {
	e.storage.Set(query.Arguments[0], query.Arguments[1])
}

func (e *Engine) processGet(ctx context.Context, query compute.Query) string {
	val, _ := e.storage.Get(query.Arguments[0])

	return val
}

func (e *Engine) processDel(ctx context.Context, query compute.Query) {
	e.storage.Del(query.Arguments[0])
}
