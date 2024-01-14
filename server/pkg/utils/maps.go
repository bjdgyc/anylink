package utils

import (
	"sync"

	cmap "github.com/orcaman/concurrent-map"
)

type IMaps interface {
	Set(key string, val any)
	Get(key string) (any, bool)
	Del(key string)
}

/**
 * 基础的Map结构
 *
 */
type BaseMap struct {
	m map[string]any
}

func (m *BaseMap) Set(key string, value any) {
	m.m[key] = value
}
func (m *BaseMap) Get(key string) (any, bool) {
	v, ok := m.m[key]
	return v, ok
}
func (m *BaseMap) Del(key string) {
	delete(m.m, key)
}

/**
 * CMap 并发结构
 *
 */
type ConcurrentMap struct {
	m cmap.ConcurrentMap
}

func (m *ConcurrentMap) Set(key string, value any) {
	m.m.Set(key, value)
}

func (m *ConcurrentMap) Get(key string) (any, bool) {
	return m.m.Get(key)
}

func (m *ConcurrentMap) Del(key string) {
	m.m.Remove(key)
}

/**
 * Map 读写结构
 *
 */
type RWLockMap struct {
	m    map[string]any
	lock sync.RWMutex
}

func (m *RWLockMap) Set(key string, value any) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.m[key] = value
}

func (m *RWLockMap) Get(key string) (any, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	v, ok := m.m[key]
	return v, ok
}

func (m *RWLockMap) Del(key string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.m, key)
}

/**
 * sync.Map 结构
 *
 */
type SyncMap struct {
	m sync.Map
}

func (m *SyncMap) Set(key string, val any) {
	m.m.Store(key, val)
}

func (m *SyncMap) Get(key string) (any, bool) {
	return m.m.Load(key)
}

func (m *SyncMap) Del(key string) {
	m.m.Delete(key)
}

func NewMap(name string, len int) IMaps {
	switch name {
	case "cmap":
		return &ConcurrentMap{m: cmap.New()}
	case "rwmap":
		m := make(map[string]any, len)
		return &RWLockMap{m: m}
	case "syncmap":
		return &SyncMap{}
	default:
		m := make(map[string]any, len)
		return &BaseMap{m: m}
	}
}
