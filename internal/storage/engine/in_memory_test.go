package engine

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/JaneJavannie/in_memory_key_value_db/internal/compute"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/configs"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/consts"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/wal"
	"github.com/stretchr/testify/require"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestEngine_ProcessCommand(t *testing.T) {
	type args struct {
		ctx   context.Context
		query compute.Query
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "set success",
			args: args{
				ctx: context.WithValue(context.Background(), consts.RequestID, uuid.New().String()),
				query: compute.Query{
					Command:   "SET",
					Arguments: []string{"hello", "world"},
				},
			},
			want: "",
		},

		{
			name: "get success",
			args: args{
				ctx: context.WithValue(context.Background(), consts.RequestID, uuid.New().String()),
				query: compute.Query{
					Command:   "GET",
					Arguments: []string{"hello", "world"},
				},
			},
			want: "world",
		},

		{
			name: "get error",
			args: args{
				ctx: context.WithValue(context.Background(), consts.RequestID, uuid.New().String()),
				query: compute.Query{
					Command:   "GET",
					Arguments: []string{"not_set", "111"},
				},
			},
			want: "",
		},

		{
			name: "del success",
			args: args{
				ctx: context.WithValue(context.Background(), consts.RequestID, uuid.New().String()),
				query: compute.Query{
					Command:   "DEL",
					Arguments: []string{"hello", "world"},
				},
			},
			want: "",
		},
	}

	storage, err := NewInMemoryStorage(nil)
	require.NoError(t, err)

	e := &Engine{
		storage: storage,
	}

	for _, tt := range tests {
		got, err := e.ProcessCommand(tt.args.ctx, tt.args.query)
		require.NoError(t, err)

		assert.Equalf(t, tt.want, got, "ProcessCommand(%v, %v)", tt.args.ctx, tt.args.query)
	}
}

func TestNewInMemoryEngine_Master(t *testing.T) {
	storage := &InMemoryStorage{}
	wal := &wal.Wal{}
	logger := slog.Default()
	cfgWal := &configs.Wal{}
	replicationType := "master"

	engine, err := NewInMemoryEngine(storage, wal, logger, cfgWal, replicationType)

	assert.NoError(t, err)
	assert.NotNil(t, engine)
	assert.Equal(t, logger, engine.logger)
	assert.Equal(t, storage, engine.storage)
	assert.True(t, engine.isWriteWal)
	assert.Equal(t, wal, engine.wal)
	assert.False(t, engine.isSlave)
}

func TestNewInMemoryEngine_Slave(t *testing.T) {
	storage := &InMemoryStorage{}
	wal := &wal.Wal{}
	logger := slog.Default()
	cfgWal := &configs.Wal{}
	replicationType := "slave"

	engine, err := NewInMemoryEngine(storage, wal, logger, cfgWal, replicationType)

	assert.NoError(t, err)
	assert.NotNil(t, engine)
	assert.Equal(t, logger, engine.logger)
	assert.Equal(t, storage, engine.storage)
	assert.False(t, engine.isWriteWal)
	assert.Equal(t, wal, engine.wal)
	assert.True(t, engine.isSlave)
}

func TestNewInMemoryEngine_NoConfig(t *testing.T) {
	storage := &InMemoryStorage{}
	wal := &wal.Wal{}
	logger := slog.Default()
	replicationType := "master"

	engine, err := NewInMemoryEngine(storage, wal, logger, nil, replicationType)

	assert.NoError(t, err)
	assert.NotNil(t, engine)
	assert.Equal(t, logger, engine.logger)
	assert.Equal(t, storage, engine.storage)
	assert.False(t, engine.isWriteWal)
	assert.Equal(t, wal, engine.wal)
	assert.False(t, engine.isSlave)
}

func TestEngine_writeWalRecord(t *testing.T) {
	ctx := context.WithValue(context.Background(), consts.RequestID, "123")
	query := compute.Query{Command: "SET", Arguments: []string{"test1", "test2"}}

	cfg := &configs.Wal{
		FlushingBatchSize:    1,
		FlushingBatchTimeout: 1 * time.Second,
		MaxSegmentSize:       "1KB",
		MaxSegmentSizeBytes:  1024,
		DataDir:              os.TempDir(),
	}

	logger := slog.Default()

	w, err := wal.NewWal(logger, cfg, "master")
	assert.NoError(t, err)

	engine := &Engine{logger: logger, wal: w}

	w.Start(cfg)

	err = engine.writeWalRecord(ctx, query)
	assert.NoError(t, err)
}
