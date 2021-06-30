package dbdata

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var autocache = cache.New(300*time.Second, 60*time.Second)

func InArrStr(arr []string, str string) bool {
	for _, d := range arr {
		if d == str {
			return true
		}
	}
	return false
}

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
	TB = 1024 * GB
	PB = 1024 * TB
)

func HumanByte(bf interface{}) string {
	var hb string
	var bAll float64
	switch bi := bf.(type) {
	case int:
		bAll = float64(bi)
	case int32:
		bAll = float64(bi)
	case uint32:
		bAll = float64(bi)
	case int64:
		bAll = float64(bi)
	case uint64:
		bAll = float64(bi)
	case float64:
		bAll = float64(bi)
	}

	switch {
	case bAll >= TB:
		hb = fmt.Sprintf("%0.2f TB", bAll/TB)
	case bAll >= GB:
		hb = fmt.Sprintf("%0.2f GB", bAll/GB)
	case bAll >= MB:
		hb = fmt.Sprintf("%0.2f MB", bAll/MB)
	case bAll >= KB:
		hb = fmt.Sprintf("%0.2f KB", bAll/KB)
	default:
		hb = fmt.Sprintf("%0.2f B", bAll)
	}

	return hb
}

func RandomRunes(length int) string {
	letterRunes := []rune("abcdefghijklmnpqrstuvwxy1234567890")

	bytes := make([]rune, length)

	for i := range bytes {
		bytes[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(bytes)
}

//SHARD_COUNT 分片数量
var SHARD_COUNT = 256

const prime32 = uint32(16777619)
const hash32 = uint32(2166136261)

//ConcurrentMap 类型的“线程”安全映射任何东西。
//为了避免锁定瓶颈，该映射被划分为几个（SHARD_COUNT）个贴图碎片。
type ConcurrentMap []*ConcurrentMapShared

//ConcurrentMapShared 任何映射的“线程”安全字符串。
type ConcurrentMapShared struct {
	items        map[string]interface{}
	sync.RWMutex // Read Write mutex, guards access to internal map.
}

//New 创建新的并发映射。
func New() ConcurrentMap {
	m := make(ConcurrentMap, SHARD_COUNT)
	for i := 0; i < SHARD_COUNT; i++ {
		m[i] = &ConcurrentMapShared{items: make(map[string]interface{}, SHARD_COUNT)}
	}
	return m
}

//GetShard 返回给定密钥下的分片
func (m ConcurrentMap) GetShard(key string) *ConcurrentMapShared {
	return m[uint(fnv32(key))%uint(SHARD_COUNT)]
}

//MSet 将普通map数据k，v插入并法map中
func (m ConcurrentMap) MSet(data map[string]interface{}) {
	for key, value := range data {
		shard := m.GetShard(key)
		shard.Lock()
		shard.items[key] = value
		shard.Unlock()
	}
}

//Set 设置指定键下的给定值。
func (m ConcurrentMap) Set(key string, value interface{}) {
	// Get map shard.
	shard := m.GetShard(key)
	shard.Lock()
	shard.items[key] = value
	shard.Unlock()
}

//UpsertCb 回调返回要插入到映射中的新元素
//它是在保持锁的情况下调用的，因此它不能
//尝试访问同一映射中的其他键，因为这会导致死锁
//去吧同步.rBlock不可重入
type UpsertCb func(exist bool, valueInMap interface{}, newValue interface{}) interface{}

//Upsert 插入或更新-使用UpsertCb更新现有元素或插入新元素
func (m ConcurrentMap) Upsert(key string, value interface{}, cb UpsertCb) (res interface{}) {
	shard := m.GetShard(key)
	shard.Lock()
	v, ok := shard.items[key]
	res = cb(ok, v, value)
	shard.items[key] = res
	shard.Unlock()
	return res
}

// SetIfAbsent 如果没有值与指定键关联，则设置指定键下的给定值。
func (m ConcurrentMap) SetIfAbsent(key string, value interface{}) bool {
	// Get map shard.
	shard := m.GetShard(key)
	shard.Lock()
	_, ok := shard.items[key]
	if !ok {
		shard.items[key] = value
	}
	shard.Unlock()
	return !ok
}

// Get 从给定键下的映射检索元素。
func (m ConcurrentMap) Get(key string) (interface{}, bool) {
	// Get shard
	shard := m.GetShard(key)
	shard.RLock()
	// Get item from shard.
	val, ok := shard.items[key]
	shard.RUnlock()
	return val, ok
}

// Count 返回映射中元素的数目。
func (m ConcurrentMap) Count() int {
	count := 0
	for i := 0; i < SHARD_COUNT; i++ {
		shard := m[i]
		shard.RLock()
		count += len(shard.items)
		shard.RUnlock()
	}
	return count
}

// Has 在指定键下查找项
func (m ConcurrentMap) Has(key string) bool {
	// Get shard
	shard := m.GetShard(key)
	shard.RLock()
	// See if element is within shard.
	_, ok := shard.items[key]
	shard.RUnlock()
	return ok
}

// Remove 从映射中移除元素。
func (m ConcurrentMap) Remove(key string) {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Lock()
	delete(shard.items, key)
	shard.Unlock()
}

//RemoveCb 是在映射.RemoveCb（）保持锁定时调用
//如果返回true，则元素将从映射中移除
type RemoveCb func(key string, v interface{}, exists bool) bool

//RemoveCb 锁定包含密钥的碎片，检索其当前值并使用这些参数调用回调
//如果回调返回true并且元素存在，它将从映射中删除它
//返回回调返回的值（即使元素不在映射中）
func (m ConcurrentMap) RemoveCb(key string, cb RemoveCb) bool {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Lock()
	v, ok := shard.items[key]
	remove := cb(key, v, ok)
	if remove && ok {
		delete(shard.items, key)
	}
	shard.Unlock()
	return remove
}

// Pop 从映射中移除元素并返回它
func (m ConcurrentMap) Pop(key string) (v interface{}, exists bool) {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Lock()
	v, exists = shard.items[key]
	delete(shard.items, key)
	shard.Unlock()
	return v, exists
}

// IsEmpty 检查映射是否为空。
func (m ConcurrentMap) IsEmpty() bool {
	return m.Count() == 0
}

// Tuple Iter&IterBuffered函数用于将两个变量包装在一个通道上，
type Tuple struct {
	Key string
	Val interface{}
}

// Iter 返回一个迭代器，该迭代器可用于for range循环。
//不推荐：使用IterBuffered（）将获得更好的性能
func (m ConcurrentMap) Iter() <-chan Tuple {
	chans := snapshot(m)
	ch := make(chan Tuple)
	go fanIn(chans, ch)
	return ch
}

// IterBuffered 返回一个缓冲迭代器，该迭代器可用于for range循环。
func (m ConcurrentMap) IterBuffered() <-chan Tuple {
	chans := snapshot(m)
	total := 0
	for _, c := range chans {
		total += cap(c)
	}
	ch := make(chan Tuple, total)
	go fanIn(chans, ch)
	return ch
}

//返回包含每个碎片中元素的通道数组，
//很可能是“m”的快照。
//一旦确定了每个缓冲通道的大小，它就会返回，
//在使用goroutines填充所有通道之前。
func snapshot(m ConcurrentMap) (chans []chan Tuple) {
	chans = make([]chan Tuple, SHARD_COUNT)
	wg := sync.WaitGroup{}
	wg.Add(SHARD_COUNT)
	// Foreach shard.
	for index, shard := range m {
		go func(index int, shard *ConcurrentMapShared) {
			// Foreach key, value pair.
			shard.RLock()
			chans[index] = make(chan Tuple, len(shard.items))
			wg.Done()
			for key, val := range shard.items {
				chans[index] <- Tuple{key, val}
			}
			shard.RUnlock()
			close(chans[index])
		}(index, shard)
	}
	wg.Wait()
	return chans
}

// fanIn 将元素从chans读入channel`out`
func fanIn(chans []chan Tuple, out chan Tuple) {
	wg := sync.WaitGroup{}
	wg.Add(len(chans))
	for _, ch := range chans {
		go func(ch chan Tuple) {
			for t := range ch {
				out <- t
			}
			wg.Done()
		}(ch)
	}
	wg.Wait()
	close(out)
}

// Items 以 map[string]interface{} 返回所有项
func (m ConcurrentMap) Items() map[string]interface{} {
	tmp := make(map[string]interface{}, m.Count())

	// Insert items to temporary map.
	ch := m.IterBuffered()
	for item := range ch {
		tmp[item.Key] = item.Val
	}

	return tmp
}

//IterCb 迭代器回调，为在中找到的每个键、值调用
//地图。对给定碎片的所有调用都保持RLock
//因此回调sess是一个碎片的一致视图，
//但不是穿过碎片
type IterCb func(key string, v interface{})

//IterCb 基于回调的迭代器，最便宜的读取方式
//映射中的所有元素。
func (m ConcurrentMap) IterCb(fn IterCb) {
	for idx := range m {
		shard := (m)[idx]
		shard.RLock()
		for key, value := range shard.items {
			fn(key, value)
		}
		shard.RUnlock()
	}
}

// Keys 以 []string 形式返回所有键
func (m ConcurrentMap) Keys() []string {
	count := m.Count()
	ch := make(chan string, count)
	go func() {
		// Foreach shard.
		wg := sync.WaitGroup{}
		wg.Add(SHARD_COUNT)
		for _, shard := range m {
			go func(shard *ConcurrentMapShared) {
				// Foreach key, value pair.
				shard.RLock()
				for key := range shard.items {
					ch <- key
				}
				shard.RUnlock()
				wg.Done()
			}(shard)
		}
		wg.Wait()
		close(ch)
	}()

	// Generate keys
	keys := make([]string, 0, count)
	for k := range ch {
		keys = append(keys, k)
	}
	return keys
}

//MarshalJSON 转json 字符切片

func (m ConcurrentMap) MarshalJSON() ([]byte, error) {
	// Create a temporary map, which will hold all item spread across shards.
	tmp := make(map[string]interface{}, m.Count())

	// Insert items to temporary map.
	for item := range m.IterBuffered() {
		tmp[item.Key] = item.Val
	}

	return json.Marshal(tmp)
}

func fnv32(key string) uint32 {
	hash := hash32
	//const prime32 = uint32(16777619)
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
	//return crc32.ChecksumIEEE([]byte(key))
}

// Concurrent map uses Interface{} as its value, therefor JSON Unmarshal
// will probably won't know which to type to unmarshal into, in such case
// we'll end up with a value of type map[string]interface{}, In most cases this isn't
// out value type, this is why we've decided to remove this functionality.

// func (m *ConcurrentMap) UnmarshalJSON(b []byte) (err error) {
// 	// Reverse process of Marshal.

// 	tmp := make(map[string]interface{})

// 	// Unmarshal into a single map.
// 	if err := json.Unmarshal(b, &tmp); err != nil {
// 		return nil
// 	}

// 	// foreach key,value pair in temporary map insert into our concurrent map.
// 	for key, val := range tmp {
// 		m.Set(key, val)
// 	}
// 	return nil
// }
