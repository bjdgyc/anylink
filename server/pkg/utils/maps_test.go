package utils

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	NumOfReader = 200
	NumOfWriter = 100
)

func TestMaps(t *testing.T) {
	assert := assert.New(t)
	var ipAuditMap IMaps
	key := "one"
	value := 100

	testMapData := map[string]int{"basemap": 512, "cmap": 0, "rwmap": 512, "syncmap": 0}
	for name, len := range testMapData {
		ipAuditMap = NewMap(name, len)
		ipAuditMap.Set(key, value)
		v, ok := ipAuditMap.Get(key)
		assert.Equal(v.(int), value)
		assert.True(ok)
		ipAuditMap.Del(key)
		v, ok = ipAuditMap.Get(key)
		assert.Nil(v)
		assert.False(ok)
	}
}

func benchmarkMap(b *testing.B, hm IMaps) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for i := 0; i < NumOfWriter; i++ {
			wg.Add(1)
			go func() {
				for i := 0; i < 100; i++ {
					hm.Set(strconv.Itoa(i), i*i)
					hm.Set(strconv.Itoa(i), i*i)
					hm.Del(strconv.Itoa(i))
				}
				wg.Done()
			}()
		}
		for i := 0; i < NumOfReader; i++ {
			wg.Add(1)
			go func() {
				for i := 0; i < 100; i++ {
					hm.Get(strconv.Itoa(i))
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkMaps(b *testing.B) {
	b.Run("RW map", func(b *testing.B) {
		myMap := NewMap("rwmap", 512)
		benchmarkMap(b, myMap)
	})
	b.Run("Concurrent map", func(b *testing.B) {
		myMap := NewMap("cmap", 0)
		benchmarkMap(b, myMap)
	})
	b.Run("Sync map", func(b *testing.B) {
		myMap := NewMap("syncmap", 0)
		benchmarkMap(b, myMap)
	})
}
