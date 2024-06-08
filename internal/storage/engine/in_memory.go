package engine

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/JaneJavannie/in_memory_key_value_db/internal/compute"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/configs"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/consts"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/wal"
)

type Engine struct {
	logger  *slog.Logger
	storage *InMemoryStorage

	isWriteWal bool
	wal        *wal.Wal
}

func NewInMemoryEngine(storage *InMemoryStorage, wal *wal.Wal, logger *slog.Logger, cfgWal configs.Wal) (*Engine, error) {
	e := Engine{
		logger:     logger,
		storage:    storage,
		isWriteWal: cfgWal.IsWriteToWal,
		wal:        wal,
	}

	return &e, nil
}

func (e *Engine) ProcessCommand(ctx context.Context, query compute.Query) (string, error) {
	slog.Debug("processing command", consts.RequestID, ctx.Value(consts.RequestID).(string), "command", query.Command)

	queryResult := ""
	var err error

	switch query.Command {
	case consts.CommandSet:
		err = e.processSet(ctx, query)

	case consts.CommandGet:
		queryResult = e.processGet(ctx, query)

	case consts.CommandDel:
		err = e.processDel(ctx, query)
	}

	return queryResult, err
}

func (e *Engine) processSet(ctx context.Context, query compute.Query) error {
	var err error
	if e.isWriteWal {
		err = e.writeWalRecord(ctx, query)
		if err != nil {
			return fmt.Errorf("write wal record: %w", err)
		}
	}

	e.storage.Set(query.Arguments[0], query.Arguments[1])

	return nil
}

func (e *Engine) processGet(_ context.Context, query compute.Query) string {
	val, _ := e.storage.Get(query.Arguments[0])

	return val
}

func (e *Engine) processDel(ctx context.Context, query compute.Query) error {
	var err error

	if e.isWriteWal {
		err = e.writeWalRecord(ctx, query)
		if err != nil {
			return fmt.Errorf("write wal record: %w", err)
		}
	}

	e.storage.Del(query.Arguments[0])

	return nil
}

func (e *Engine) writeWalRecord(ctx context.Context, query compute.Query) error {
	id := ctx.Value(consts.RequestID).(string)
	log := wal.Log{
		ID:    id,
		Query: query,
	}

	e.logger.Debug("WAIT WAL")
	err := e.wal.WriteLog(ctx, log)
	e.logger.Debug("WAIT DONE")

	return err
}
