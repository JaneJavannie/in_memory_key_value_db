package engine

import (
	"sync"
)

type kvStorage struct {
	mu sync.Mutex
	m  map[string]string
}

func (s *kvStorage) set(key string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.m[key] = value
}

func (s *kvStorage) get(key string) (string, bool) {
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
