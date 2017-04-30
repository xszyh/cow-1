package proxy

// Automatically generated file; DO NOT EDIT

import (
	"sync"
)

// proxySafeMap is a thread-safe map mapping from string to bool.
type proxySafeMap struct {
	m    map[string]bool
	lock sync.RWMutex
}

// NewproxySafeMap returns a new proxySafeMap.
func NewproxySafeMap(m map[string]bool) *proxySafeMap {
	if m == nil {
		m = make(map[string]bool)
	}
	return &proxySafeMap{
		m: m,
	}

}

// Get returns a point of bool, it returns nil if not found.
func (s *proxySafeMap) Get(k string) (bool, bool) {
	s.lock.RLock()
	v, ok := s.m[k]
	s.lock.RUnlock()
	return v, ok
}

// Set sets value v to key k in the map.
func (s *proxySafeMap) Set(k string, v bool) {
	s.lock.Lock()
	s.m[k] = v
	s.lock.Unlock()
}

// Update updates value v to key k, returns false if k not found.
func (s *proxySafeMap) Update(k string, v bool) bool {
	s.lock.Lock()
	_, ok := s.m[k]
	if !ok {
		s.lock.Unlock()
		return false
	}
	s.m[k] = v
	s.lock.Unlock()
	return true
}

// Delete deletes a key in the map.
func (s *proxySafeMap) Delete(k string) {
	s.lock.Lock()
	delete(s.m, k)
	s.lock.Unlock()
}

// Dup duplicates the map to a new struct.
func (s *proxySafeMap) Dup() *proxySafeMap {
	newMap := NewproxySafeMap(nil)
	s.lock.Lock()
	for k, v := range s.m {
		newMap.m[k] = v
	}
	s.lock.Unlock()
	return newMap
}
