package utils

import (
	"sync"

	cmap "github.com/orcaman/concurrent-map"
)

type IMaps interface {
	Set(key string, val interface{})
	Get(key string) (interface{}, bool)
	Del(key string)
}

/**
 * 基础的Map结构
 *
 */
type BaseMap struct {
	m map[string]interface{}
}

func (m *BaseMap) Set(key string, value interface{}) {
	m.m[key] = value
}
func (m *BaseMap) Get(key string) (interface{}, bool) {
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

func (m *ConcurrentMap) Set(key string, value interface{}) {
	m.m.Set(key, value)
}

func (m *ConcurrentMap) Get(key string) (interface{}, bool) {
	v, ok := m.m.Get(key)
	return v, ok
}

func (m *ConcurrentMap) Del(key string) {
	m.m.Remove(key)
}

/**
 * Map 读写结构
 *
 */
type RWLockMap struct {
	m    map[string]interface{}
	lock sync.RWMutex
}

func (m *RWLockMap) Set(key string, value interface{}) {
	m.lock.Lock()
	m.m[key] = value
	m.lock.Unlock()
}

func (m *RWLockMap) Get(key string) (interface{}, bool) {
	m.lock.RLock()
	v, ok := m.m[key]
	m.lock.RUnlock()
	return v, ok
}

func (m *RWLockMap) Del(key string) {
	m.lock.Lock()
	delete(m.m, key)
	m.lock.Unlock()
}

/**
 * sync.Map 结构
 *
 */
type SyncMap struct {
	m sync.Map
}

func (m *SyncMap) Set(key string, val interface{}) {
	m.m.Store(key, val)
}

func (m *SyncMap) Get(key string) (interface{}, bool) {
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
		m := make(map[string]interface{}, len)
		return &RWLockMap{m: m}
	case "syncmap":
		return &SyncMap{}
	default:
		m := make(map[string]interface{}, len)
		return &BaseMap{m: m}
	}
}
