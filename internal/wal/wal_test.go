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
	"github.com/stretchr/testify/assert"
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

	err := WriteRecord(dataDir, filename, walRecords)
	if err != nil {
		t.Errorf("Error writing record: %v", err)
	}
}

func TestWriteRecord_OpenFileError(t *testing.T) {
	dataDir := "/path/to/data"
	filename := "test.txt"
	walRecords := bytes.NewBufferString("test data")

	err := WriteRecord(dataDir, filename, *walRecords)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "open file")
}

func TestWriteRecord_WriteFileError(t *testing.T) {
	dataDir := "/path/to/data"
	filename := "test.txt"
	walRecords := bytes.NewBufferString("test data")

	err := WriteRecord(dataDir, filename, *walRecords)

	assert.Error(t, err)
}

func TestBuildWalRecordsFromMap(t *testing.T) {
	logs := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	walRecords := buildWalRecordsFromMap(logs)

	assert.Contains(t, walRecords.String(), "SET key1 value1 \n", "SET key2 value2 \n")
}

func TestRemoveFileWithRetries(t *testing.T) {
	dataDir := "./"
	fileName := "testfile"

	file, err := os.CreateTemp(dataDir, fileName)
	if err != nil {
		t.Fatalf("Error creating temporary file: %v", err)
	}

	err = removeFileWithRetries(dataDir, file.Name())

	assert.NoError(t, err)
}
