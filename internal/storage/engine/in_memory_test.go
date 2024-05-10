package engine

import (
	"context"
	"fmt"
	"testing"

	"in_memory_key_value_db/internal/compute"
	"in_memory_key_value_db/internal/consts"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestEngine_ProcessCommand(t *testing.T) {
	type args struct {
		ctx   context.Context
		query compute.Query
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
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
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err, "expected no error")
			},
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
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err, "expected no error")
			},
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
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, "expected error")
			},
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
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err, "expected no error")
			},
		},
	}

	e := &Engine{
		storage: InitMemoryStorage(),
	}

	for _, tt := range tests {
		got, err := e.ProcessCommand(tt.args.ctx, tt.args.query)
		if !tt.wantErr(t, err, fmt.Sprintf("ProcessCommand(%v, %v)", tt.args.ctx, tt.args.query)) {
			return
		}
		assert.Equalf(t, tt.want, got, "ProcessCommand(%v, %v)", tt.args.ctx, tt.args.query)
	}
}
