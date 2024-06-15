package wal

import (
	"bytes"
	"log/slog"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/JaneJavannie/in_memory_key_value_db/internal/compute"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/configs"
)

func TestStart(t *testing.T) {
	dataDir := os.TempDir()

	w := &Wal{
		batch: []Log{
			{
				ID: strconv.Itoa(1), Query: compute.Query{Command: "SET", Arguments: []string{"a", "b"}},
			},
			{
				ID: strconv.Itoa(2), Query: compute.Query{Command: "DEL", Arguments: []string{"a", "b"}},
			},
		},
		dataDir:               dataDir,
		maxLogFileSegmentSize: 1024,
		batchTimeout:          5 * time.Second,
		batchSize:             2,
		operations:            make(chan Log),
		logger:                slog.Default(),
	}

	go w.Start(&configs.Wal{})
}

func TestFlushRecords(t *testing.T) {
	dataDir := os.TempDir()

	w := &Wal{
		batch:                 []Log{{ID: strconv.Itoa(1), Query: compute.Query{Command: "SET", Arguments: []string{"aa", "bb"}}}},
		dataDir:               dataDir,
		maxLogFileSegmentSize: 1024,
	}

	err := w.flushRecords()
	if err != nil {
		t.Errorf("Error flushing records: %v", err)
	}
}

func TestWriteRecord(t *testing.T) {
	dataDir := os.TempDir()
	//defer os.RemoveAll(dataDir)

	filename := "test.txt"
	walRecords := bytes.Buffer{}
	walRecords.WriteString("Test data")

	err := writeRecord(dataDir, filename, walRecords)
	if err != nil {
		t.Errorf("Error writing record: %v", err)
	}
}
