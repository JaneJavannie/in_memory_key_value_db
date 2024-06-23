package text

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTcpServer_Start(t *testing.T) {
	server := &TcpServer{address: "localhost:8080"}
	server.log = slog.Default()

	err := server.Start()

	assert.NoError(t, err)
}
