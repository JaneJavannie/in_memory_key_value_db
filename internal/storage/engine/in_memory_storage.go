package engine

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/JaneJavannie/in_memory_key_value_db/internal/configs"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/consts"
	"github.com/cespare/xxhash/v2"
)

const (
	bucketCount     = 256
	defaultKeyCount = 8
)

type InMemoryStorage struct {
	data [bucketCount]*kvStorage
}

func NewInMemoryStorage(cfg *configs.Config) (*InMemoryStorage, error) {
	c := &InMemoryStorage{}

	for i := 0; i < bucketCount; i++ {
		c.data[i] = &kvStorage{
			mu: sync.Mutex{},
			m:  make(map[string]string, defaultKeyCount),
		}
	}

	replication := cfg.Replication

	if cfg.Wal == nil || replication == nil {
		return c, nil
	}

	dataDir := ""

	if cfg.Wal != nil {
		dataDir = cfg.Wal.DataDir
	}

	if replication != nil && replication.Type == consts.ReplicationTypeSlave {
		dataDir = replication.ReplicatedDataDir
	}

	err := c.loadWal(dataDir)
	if err != nil {
		return nil, fmt.Errorf("load WAL: %v", err)
	}

	return c, nil
}

func (c *InMemoryStorage) Set(key string, value string) {
	hash := getHash(key, len(c.data))
	bucket := c.data[hash]

	bucket.set(key, value)
}

func (c *InMemoryStorage) Get(key string) (string, bool) {
	hash := getHash(key, len(c.data))
	bucket := c.data[hash]

	value, ok := bucket.get(key)
	if !ok {
		return "", false
	}

	return value, true
}

func (c *InMemoryStorage) Del(key string) {
	hash := getHash(key, len(c.data))
	bucket := c.data[hash]

	bucket.del(key)
}

func (c *InMemoryStorage) loadWal(dir string) error {
	logs, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read dir %s: %v", dir, err)
	}

	for _, log := range logs {
		if log.IsDir() {
			continue
		}

		file, err := os.Open(filepath.Join(dir, log.Name()))
		if err != nil {
			return fmt.Errorf("error opening file: %v\n", err)
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			entries := strings.Split(scanner.Text(), " ")

			args := make([]string, 0)
			args = append(args, entries[2:]...)

			command := entries[1]

			switch command {
			case consts.CommandSet:
				c.Set(args[0], args[1])

			case consts.CommandDel:
				c.Del(args[0])

			default:
				return fmt.Errorf("unknown command: %s", command)
			}
		}

	}

	return nil
}

func getHash(key string, bucketCount int) int {
	h := xxhash.Sum64String(key) % uint64(bucketCount)

	return int(h)
}
