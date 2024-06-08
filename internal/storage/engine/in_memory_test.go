package engine

import (
	"context"
	"testing"

	"github.com/JaneJavannie/in_memory_key_value_db/internal/compute"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/consts"
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

	storage, err := NewInMemoryStorage(false, "")
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
