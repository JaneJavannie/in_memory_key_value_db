package engine

import (
	"sync"

	"github.com/cespare/xxhash/v2"
)

const (
	bucketCount     = 256
	defaultKeyCount = 8
)

type InMemoryStorage struct {
	data [bucketCount]*kvStorage
}

func InitMemoryStorage() *InMemoryStorage {
	c := &InMemoryStorage{}

	for i := 0; i < bucketCount; i++ {
		c.data[i] = &kvStorage{
			mu: sync.Mutex{},
			m:  make(map[string]string, defaultKeyCount),
		}
	}

	return c
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

func getHash(key string, bucketCount int) int {
	h := xxhash.Sum64String(key) % uint64(bucketCount)

	return int(h)
}
