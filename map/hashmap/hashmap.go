package hashmap

import (
	"runtime"
	"unsafe"
)

type HashMap[K comparable, V any] struct {
	shardCount int
	shards     []ShardMap[K, V]
	hasher     func(key unsafe.Pointer, seed uintptr) uintptr
	seed       uintptr
}

type HashMapOption[K comparable, V any] func(*HashMap[K, V])

func WithShardCount[K comparable, V any](shardCount int) HashMapOption[K, V] {
	return func(m *HashMap[K, V]) {
		m.shardCount = nextPowerOfTwo(shardCount)
	}
}

func NewHashMap[K comparable, V any](options ...HashMapOption[K, V]) *HashMap[K, V] {
	m := &HashMap[K, V]{
		shardCount: nextPowerOfTwo(runtime.GOMAXPROCS(0) * 16),
		hasher:     getRuntimeHasher[K](),
		seed:       uintptr(fastrand64()),
	}
	for _, option := range options {
		option(m)
	}
	m.shards = make([]ShardMap[K, V], m.shardCount)
	for i := 0; i < m.shardCount; i++ {
		m.shards[i] = NewShardMap[K, V]()
	}
	return m
}

func (m *HashMap[K, V]) getShard(key K) *ShardMap[K, V] {
	hash := uint32(m.hasher(noescape(unsafe.Pointer(&key)), m.seed))
	return &m.shards[int(hash)%m.shardCount]
}

func (m *HashMap[K, V]) Set(k K, v V) {
	m.getShard(k).Set(k, v)
}

func (m *HashMap[K, V]) Get(k K) (V, bool) {
	return m.getShard(k).Get(k)
}

func (m *HashMap[K, V]) Delete(k K) {
	m.getShard(k).Delete(k)
}
