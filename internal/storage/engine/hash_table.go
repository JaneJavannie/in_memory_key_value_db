package engine

import (
	"sync"

	"github.com/cespare/xxhash/v2"
)

const (
	bucketCount     = 256
	defaultKeyCount = 8
)

type Value struct {
	val string
}

type kvStorage struct {
	mu sync.Mutex
	m  map[string]*Value
}

type InMemoryStorage struct {
	data [bucketCount]*kvStorage
}

func InitMemoryStorage() *InMemoryStorage {
	c := &InMemoryStorage{}

	for i := 0; i < bucketCount; i++ {
		c.data[i] = &kvStorage{
			mu: sync.Mutex{},
			m:  make(map[string]*Value, defaultKeyCount),
		}
	}

	return c
}

func (s *kvStorage) set(key string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.m[key] = &Value{
		val: value,
	}
}

func (s *kvStorage) get(key string) (*Value, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.m[key]

	return val, ok
}

func (s *kvStorage) del(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.m, key)
}

func (c *InMemoryStorage) Set(key string, value string) {
	hash := int(xxhash.Sum64String(key)) % len(c.data)
	bucket := c.data[hash]

	bucket.set(key, value)
}

func (c *InMemoryStorage) Get(key string) (string, bool) {
	hash := int(xxhash.Sum64String(key)) % len(c.data)
	bucket := c.data[hash]

	value, ok := bucket.get(key)
	if !ok {
		return "", false
	}

	return value.val, true
}

func (c *InMemoryStorage) Del(key string) {
	hash := int(xxhash.Sum64String(key)) % len(c.data)
	bucket := c.data[hash]

	bucket.del(key)
}
