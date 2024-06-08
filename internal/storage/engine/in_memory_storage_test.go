package engine

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadWal(t *testing.T) {
	dir := "../../../wal_logs/wal"
	c, err := NewInMemoryStorage(true, dir)
	require.NoError(t, err)

	res, _ := c.Get("test_val")

	require.Equal(t, res, "test123")
}
