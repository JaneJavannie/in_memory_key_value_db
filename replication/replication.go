package replication

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/JaneJavannie/in_memory_key_value_db/internal/configs"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/consts"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/protocol/text"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/storage/engine"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/wal"
)

var connErr = errors.New("connection error")

type Replication struct {
	replicationType string
	masterAddress   string
	syncInterval    time.Duration
	walDir          string
	storage         *engine.InMemoryStorage
	client          *text.Client
	server          *text.TcpServer
	logger          *slog.Logger
}

type replicatedWal struct {
	FileName string
	Records  []byte
}

func NewReplication(cfg *configs.Config, client *text.Client, server *text.TcpServer, storage *engine.InMemoryStorage, logger *slog.Logger) (*Replication, error) {
	replication := cfg.Replication
	replicationType := replication.Type

	replica := &Replication{
		replicationType: replicationType,
		masterAddress:   replication.MasterAddress,
		syncInterval:    replication.SyncInterval,
		walDir:          replication.ReplicatedDataDir,
		storage:         storage,
		client:          client,
		server:          server,
		logger:          logger,
	}

	if cfg.Wal != nil && replicationType == consts.ReplicationTypeMaster {
		replica.walDir = cfg.Wal.DataDir
	}

	return replica, nil
}

func (r *Replication) Start(ctx context.Context, syncInterval time.Duration) error {
	switch r.replicationType {
	case consts.ReplicationTypeSlave:
		err := r.startSlave(ctx, syncInterval)
		if err != nil {
			return fmt.Errorf("start slave: %w", err)
		}

	case consts.ReplicationTypeMaster:
		err := r.startMaster()
		if err != nil {
			return fmt.Errorf("start master: %w", err)
		}

	default:
		return fmt.Errorf("replication type %s not supported", r.replicationType)
	}

	return nil
}

func (r *Replication) startSlave(ctx context.Context, syncInterval time.Duration) error {
	ticker := time.NewTicker(syncInterval)
	defer ticker.Stop()

	err := r.client.Connect(ctx)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			r.client.Close()
			return nil

		case <-ticker.C:
			err := r.GetWalsFromMaster(ctx)
			if err != nil {
				if errors.Is(err, connErr) {
					err = r.client.Connect(ctx)
					if err != nil {
						r.logger.Error("error", "connect", err)
					}
				} else {
					r.logger.Error("error", "get master wals", err)
				}
			}
		}
	}

}

type walMessage struct {
	WALs []replicatedWal
	Err  error
}

func (r *Replication) startMaster() error {
	r.server.SetOnReceive(func(ctx context.Context, request string) string {
		wals, err := r.SendWalsToReplica(ctx, request)
		if err != nil {
			r.logger.Error("send wal to replica", "err", err)
			return ""
		}

		res := walMessage{
			WALs: wals,
			Err:  err,
		}

		resp, err := json.Marshal(res)
		if err != nil {
			res = walMessage{Err: err}
			resp, err = json.Marshal(res)
			if err != nil {
				r.logger.Error("marshal response", "err", err)
				return ""
			}
		}

		return string(resp)
	})

	err := r.server.Start()
	if err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	return nil
}

func (r *Replication) GetWalsFromMaster(ctx context.Context) error {
	// read wals that were already copied from master
	dirEntries, err := os.ReadDir(r.walDir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("read dir: %w", err)
		}

		err = os.MkdirAll(r.walDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("create dir: %w", err)
		}
	}

	sort.Slice(dirEntries, func(i, j int) bool {
		return dirEntries[i].Name() < dirEntries[j].Name()
	})

	latest := ""
	if len(dirEntries) > 0 {
		latest = dirEntries[len(dirEntries)-1].Name()
	}

	encoded, err := r.client.Send(ctx, latest)
	if err != nil {
		return fmt.Errorf("%w: send request to get new wals from master: %w", connErr, err)
	}

	walmessage := new(walMessage)

	err = json.Unmarshal([]byte(encoded), walmessage)
	if err != nil {
		return fmt.Errorf("unmarshal new wals: %w", err)
	}
	if walmessage.Err != nil {
		return fmt.Errorf("get new wals: %w", walmessage.Err)
	}

	for _, w := range walmessage.WALs {
		buf := bytes.NewBuffer(w.Records)

		err := wal.WriteRecord(r.walDir, w.FileName, *buf)
		if err != nil {
			return fmt.Errorf("write record: %w", err)
		}

		// apply new records
		reader := bytes.NewReader(w.Records)
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			entries := strings.Split(scanner.Text(), " ")

			args := make([]string, 0)
			args = append(args, entries[2:]...)

			id := entries[0]
			command := entries[1]

			r.logger.Debug("processing log id: %v", id)

			switch command {
			case consts.CommandSet:
				r.storage.Set(args[0], args[1])

			case consts.CommandDel:
				r.storage.Del(args[0])

			default:
				return fmt.Errorf("unknown command: %s", command)
			}
		}
	}

	return nil
}

func (r *Replication) SendWalsToReplica(ctx context.Context, latestReplicatedEntry string) ([]replicatedWal, error) {
	dirEntries, err := os.ReadDir(r.walDir)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}

	sort.Slice(dirEntries, func(i, j int) bool {
		return dirEntries[i].Name() < dirEntries[j].Name()
	})

	walsToSend := make([]replicatedWal, 0)

	if len(dirEntries) == 0 {
		return nil, nil
	}

	latest := dirEntries[len(dirEntries)-1]

	// nothing to send
	if latestReplicatedEntry == latest.Name() {
		return nil, nil
	}

	// collect wal entries from the end till encounter already sent latestReplicatedEntry
	for i := len(dirEntries) - 1; i >= 0; i-- {
		fileName := dirEntries[i].Name()

		path := filepath.Join(r.walDir, fileName)

		if latestReplicatedEntry == fileName {
			break
		}

		fileBytes, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read file %s: %w", fileName, err)
		}

		walsToSend = append(walsToSend, replicatedWal{
			FileName: fileName,
			Records:  fileBytes,
		})
	}

	return walsToSend, nil
}
