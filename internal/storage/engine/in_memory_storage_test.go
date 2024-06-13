package engine

import (
	"testing"

	"github.com/JaneJavannie/in_memory_key_value_db/internal/configs"
	"github.com/stretchr/testify/require"
)

func TestLoadWal(t *testing.T) {
	dir := "../../../wal_logs/wal"
	c, err := NewInMemoryStorage(&configs.Wal{DataDir: dir})
	require.NoError(t, err)

	res, _ := c.Get("test_val")

	require.Equal(t, res, "test123")
}
