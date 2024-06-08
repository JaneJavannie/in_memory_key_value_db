package wal

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/JaneJavannie/in_memory_key_value_db/internal/compute"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/configs"
)

type Wal struct {
	logger *slog.Logger

	batchSize             int
	batchTimeout          time.Duration
	maxLogFileSegmentSize int
	dataDir               string

	// channel that contains users' modifying operations - set, del
	operations chan Log
	batch      []Log

	flushErr error
	xCond    *sync.Mutex
	cond     *sync.Cond

	close chan struct{}
	done  sync.WaitGroup
}

type Log struct {
	ID    string
	Query compute.Query
}

func NewWal(logger *slog.Logger, cfg configs.Wal) (*Wal, error) {
	if !cfg.IsWriteToWal {
		return &Wal{}, nil
	}

	wal := &Wal{
		logger: logger,

		batchSize:             cfg.FlushingBatchSize,
		batchTimeout:          cfg.FlushingBatchTimeout,
		maxLogFileSegmentSize: cfg.MaxSegmentSizeBytes,
		dataDir:               cfg.DataDir,

		operations: make(chan Log, 1), // client writes a value, and waits for its acknowledgment
		xCond:      &sync.Mutex{},
	}
	wal.cond = sync.NewCond(wal.xCond)

	if _, err := os.Stat(wal.dataDir); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(wal.dataDir, 0755)
			if err != nil {
				return nil, fmt.Errorf("mkdir all: %w", err)
			}
		} else {
			return nil, fmt.Errorf("os stat data dir: %w", err)
		}
	}

	return wal, nil
}

func (w *Wal) Start(isWriteWal bool) {
	if !isWriteWal {
		return
	}

	w.done.Add(1)

	go func() {
		defer w.done.Done()

		t := time.NewTimer(w.batchTimeout)
		defer t.Stop()

		for {
			flush := false

			select {
			case <-t.C:
				w.logger.Debug("CASE TIMER")

				flush = true

			case wl := <-w.operations:
				w.logger.Debug("CASE WAL")

				if len(w.batch) == 0 {
					w.logger.Debug("t.Reset()")

					t.Reset(w.batchTimeout)
				}

				w.batch = append(w.batch, wl)

				if len(w.batch) == w.batchSize {
					flush = true
				}
			}

			if flush {
				w.logger.Debug("FLUSH")

				err := w.flushRecords()
				if err != nil {
					w.logger.Error("flush records: %w", err)
				}

				t.Stop()
				w.batch = nil

				w.xCond.Lock()
				w.flushErr = err
				w.cond.Broadcast()
				w.xCond.Unlock()
			}
		}
	}()
}

func (w *Wal) Stop(isWriteWal bool) {
	if !isWriteWal {
		return
	}

	w.done.Wait()
}

func (w *Wal) WriteLog(_ context.Context, log Log) error {
	// do not handle ctx otherwise one request can cancel other requests
	w.operations <- log

	var err error
	w.xCond.Lock()
	w.cond.Wait()
	err = w.flushErr
	w.xCond.Unlock()

	return err
}

func (w *Wal) flushRecords() error {
	// do not write empty logs
	if len(w.batch) == 0 {
		return nil
	}

	dirEntries, err := os.ReadDir(w.dataDir)
	if err != nil {
		return fmt.Errorf("read dir: %w", err)
	}

	sort.Slice(dirEntries, func(i, j int) bool {
		return dirEntries[i].Name() < dirEntries[j].Name()
	})

	walRecords := bytes.Buffer{}

	for _, log := range w.batch {
		record := fmt.Sprintf("%s %s %s \n", log.ID, log.Query.Command, strings.Join(log.Query.Arguments, " "))
		walRecords.WriteString(record)
	}

	if len(dirEntries) > 0 {
		latest := dirEntries[len(dirEntries)-1]
		info, err := latest.Info()
		if err != nil {
			return fmt.Errorf("get info: %s: %w", latest.Name(), err)
		}

		// write to existing file if it has enough free space
		if int(info.Size())+walRecords.Len() < w.maxLogFileSegmentSize {
			err := writeRecord(w.dataDir, latest.Name(), walRecords)
			if err != nil {
				return fmt.Errorf("write record: %s: %w", latest.Name(), err)
			}

			return nil
		}
	}

	// otherwise write to a new one
	fileName := fmt.Sprintf("%s", time.Now().Format("20060102_150405"))
	err = writeRecord(w.dataDir, fileName, walRecords)
	if err != nil {
		return fmt.Errorf("write record: %s: %w", fileName, err)
	}

	return nil
}

func writeRecord(dataDir string, filename string, walRecords bytes.Buffer) error {
	path := filepath.Join(dataDir, filename)

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(walRecords.Bytes())
	if err != nil {
		return fmt.Errorf("write file: %s: %w", filename, err)
	}

	return nil
}
