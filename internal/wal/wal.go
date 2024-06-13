package wal

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/JaneJavannie/in_memory_key_value_db/internal/compute"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/configs"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/consts"
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

	wg sync.WaitGroup
}

type Log struct {
	ID    string
	Query compute.Query
}

func NewWal(logger *slog.Logger, cfg *configs.Wal, replicationType string) (*Wal, error) {
	if cfg == nil || replicationType == consts.ReplicationTypeSlave {
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

func (w *Wal) Start(wal *configs.Wal) {
	if wal == nil {
		return
	}

	w.wg.Add(1)

	go func() {
		defer w.wg.Done()

		timer := time.NewTimer(w.batchTimeout)
		defer timer.Stop()

		for {
			flush := false

			select {
			case <-timer.C:
				flush = w.handleTimerEvent()

			case wl := <-w.operations:
				flush = w.handleWALEvent(timer, wl)
			}

			if flush {
				w.handleFlush(timer)
			}
		}
	}()
}

func (w *Wal) handleTimerEvent() bool {
	w.logger.Debug("CASE TIMER")
	return true
}

func (w *Wal) handleWALEvent(timer *time.Timer, wl Log) bool {
	w.logger.Debug("CASE WAL")

	if len(w.batch) == 0 {
		w.logger.Debug("t.Reset()")
		timer.Reset(w.batchTimeout)
	}

	w.batch = append(w.batch, wl)

	return len(w.batch) == w.batchSize
}

func (w *Wal) handleFlush(timer *time.Timer) {
	w.logger.Debug("FLUSH")

	err := w.flushRecords()
	if err != nil {
		w.logger.Error("flush records: %w", err)
	}

	timer.Stop()
	w.batch = nil

	w.xCond.Lock()
	w.flushErr = err
	w.cond.Broadcast()
	w.xCond.Unlock()
}

func (w *Wal) Stop(wal *configs.Wal) {
	if wal == nil {
		return
	}

	w.wg.Wait()
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

	walRecords := buildWalRecords(w.batch)

	if err = writeWalRecords(w.dataDir, dirEntries, walRecords, w.maxLogFileSegmentSize); err != nil {
		return fmt.Errorf("write wal records: %w", err)
	}

	return nil
}

func buildWalRecords(batch []Log) bytes.Buffer {
	walRecords := bytes.Buffer{}

	for _, log := range batch {
		record := fmt.Sprintf("%s %s %s \n", log.ID, log.Query.Command, strings.Join(log.Query.Arguments, " "))
		walRecords.WriteString(record)
	}

	return walRecords
}

func writeWalRecords(dataDir string, dirEntries []fs.DirEntry, walRecords bytes.Buffer, maxLogFileSegmentSize int) error {
	if len(dirEntries) > 0 {
		latest := dirEntries[len(dirEntries)-1]
		info, err := latest.Info()
		if err != nil {
			return fmt.Errorf("get info: %s: %w", latest.Name(), err)
		}

		// write to an existing file
		if int(info.Size())+walRecords.Len() < maxLogFileSegmentSize {
			err := writeRecord(dataDir, latest.Name(), walRecords)
			if err != nil {
				return fmt.Errorf("write record: %s: %w", latest.Name(), err)
			}

			return nil
		}
	}

	// write to a new file
	fileName := fmt.Sprintf("%s", time.Now().Format("20060102_150405"))
	err := writeRecord(dataDir, fileName, walRecords)
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

	err = file.Sync()
	if err != nil {
		return fmt.Errorf("sync file by path [ %s ]:  %w", path, err)
	}

	return nil
}
