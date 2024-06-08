package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryStorage_Del(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name   string
		args   args
		actual string
	}{
		{
			name: "set char string",
			args: args{
				key:   "zzz",
				value: "aaa",
			},
			actual: "aaa",
		},
		{
			name: "set digit string",
			args: args{
				key:   "12345678910",
				value: "321",
			},
			actual: "321",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewInMemoryStorage(false, "")
			require.NoError(t, err)

			c.Set(tt.args.key, tt.args.value)

			c.Del(tt.args.key)

			_, ok := c.Get(tt.args.key)
			if ok {
				t.Errorf("Get() got = %v, want %v", ok, false)
			}
		})
	}
}

func TestInMemoryStorage_Get(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name   string
		args   args
		actual string
	}{
		{
			name: "set char string",
			args: args{
				key:   "zzz",
				value: "aaa",
			},
			actual: "aaa",
		},
		{
			name: "set digit string",
			args: args{
				key:   "12345678910",
				value: "321",
			},
			actual: "321",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewInMemoryStorage(false, "")
			require.NoError(t, err)

			c.Set(tt.args.key, tt.args.value)

			got, ok := c.Get(tt.args.key)
			if !ok {
				t.Errorf("Get() got = %v, want %v", ok, true)
			}

			assert.Equal(t, got, tt.actual)
		})
	}
}

func TestInMemoryStorage_Set(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name   string
		args   args
		actual string
	}{
		{
			name: "set char string",
			args: args{
				key:   "zzz_abc",
				value: "aaa_bbb",
			},
			actual: "aaa_bbb",
		},
		{
			name: "set digit string",
			args: args{
				key:   "12345678910",
				value: "321",
			},
			actual: "321",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewInMemoryStorage(false, "")
			require.NoError(t, err)

			c.Set(tt.args.key, tt.args.value)

			got, ok := c.Get(tt.args.key)
			if !ok {
				t.Errorf("Get() got = %v, want %v", ok, true)
			}

			assert.Equal(t, got, tt.actual)
		})
	}
}
