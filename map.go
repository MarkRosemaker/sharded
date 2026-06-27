package sharded

import "sync"

const (
	// numShards is the number of independent map+lock partitions.
	// It must be a power of two so the shard index is a cheap bitmask
	// (h & (numShards-1)) of the key hash.
	numShards = 1 << 8
	// cacheLineSize is used for padding. 64 bytes is a common CPU cache line size;
	// the exact value is not critical.
	cacheLineSize = 2 << 5
)

// HashFunc returns a 64-bit hash of a key. It should have good distribution
// and be fast. FNV-1a is used by default for strings.
type HashFunc[K comparable] func(K) uint64

// Map is a sharded concurrent map. It partitions the key space across
// multiple independent maps, each protected by its own mutex.
//
// Keys are routed to a shard via hashing. This design scales well with
// core count and provides high throughput under uniform access patterns.
// Skewed (Zipfian) workloads can cause imbalance as hot keys collide on
// fewer shards.
type Map[K comparable, V any] struct {
	shards [numShards]*shard[K, V]
	hash   HashFunc[K]
}

// shard holds one partition of the map.
type shard[K comparable, V any] struct {
	mu sync.Mutex
	m  map[K]V
	// pad keeps each shard's mutex on its own cache line to avoid
	// false sharing with neighboring shards.
	_ [cacheLineSize]byte
}

// NewMap creates a new sharded map using the provided hash function.
func NewMap[K comparable, V any](hf HashFunc[K]) *Map[K, V] {
	s := &Map[K, V]{hash: hf}
	for i := range s.shards {
		s.shards[i] = &shard[K, V]{m: map[K]V{}}
	}
	return s
}

// NewStringMap creates a sharded map specialized for string keys
// using an inlined FNV-1a hash.
func NewStringMap[V any]() *Map[string, V] {
	return NewMap[string, V](fnv1a)
}

// fnv1a is an inlined FNV-1a hash optimized for string keys.
// It avoids allocations compared to hash/fnv.
func fnv1a(s string) uint64 {
	const (
		offset uint64 = 14695981039346656037
		prime  uint64 = 1099511628211
	)

	h := offset
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= prime
	}

	return h
}

// shardFor returns the shard responsible for the given key.
func (m *Map[K, V]) shardFor(key K) *shard[K, V] {
	return m.shards[m.hash(key)&(numShards-1)]
}

// Get returns the value for a key and whether it exists.
func (m *Map[K, V]) Get(key K) (V, bool) {
	p := m.shardFor(key)
	p.mu.Lock()
	v, ok := p.m[key]
	p.mu.Unlock() // explicit, not defer (+8% here, see article)
	return v, ok
}

// Set stores or overwrites a key-value pair.
func (m *Map[K, V]) Set(key K, value V) {
	p := m.shardFor(key)
	p.mu.Lock()
	p.m[key] = value
	p.mu.Unlock()
}

// Delete removes a key if it exists.
func (m *Map[K, _]) Delete(key K) {
	sh := m.shardFor(key)
	sh.mu.Lock()
	delete(sh.m, key)
	sh.mu.Unlock()
}

// Len returns the total number of items in the map.
// It acquires all locks, so it is relatively expensive.
func (m *Map[_, _]) Len() int {
	n := 0
	for _, sh := range m.shards {
		sh.mu.Lock()
		n += len(sh.m)
		sh.mu.Unlock()
	}
	return n
}

// Range calls f for each key-value pair. Stops if yield returns false.
// It holds one shard lock at a time.
func (m *Map[K, V]) Range(yield func(K, V) bool) {
	for _, sh := range m.shards {
		sh.mu.Lock()
		for k, v := range sh.m {
			if !yield(k, v) {
				sh.mu.Unlock()
				return
			}
		}
		sh.mu.Unlock()
	}
}

// Clear removes all entries from the map.
func (m *Map[K, V]) Clear() {
	for _, sh := range m.shards {
		sh.mu.Lock()
		clear(sh.m)
		sh.mu.Unlock()
	}
}
