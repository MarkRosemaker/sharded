package sharded

import "sync"

const (
	// numShards is the number of independent map+lock partitions. It must be a
	// power of two so the shard index is a cheap bitmask of the key hash.
	numShards = 2 << 7
	// 64 bytes is a common cache-line size; the exact value is not load-bearing.
	cacheLineSize = 2 << 5
)

type shard struct {
	mu sync.Mutex
	m  map[string]string
	// pad keeps each shard's mutex on its own cache line so that locking
	// one shard does not cause false sharing with a neighbor.
	_ [cacheLineSize]byte
}

// Map partitions the key space across many independent maps, each with its own mutex.
// A key is routed to a shard by its hash, so operations on different shards never contend.
// This is the canonical high-throughput design: with enough shards relative to the
// number of cores, contention all but disappears under a uniform key distribution.
// Its weakness is a skewed (Zipfian) workload, where hot keys concentrate on a few shards.
type Map struct{ shards [numShards]*shard }

func NewMap() *Map {
	s := &Map{}
	for i := range s.shards {
		s.shards[i] = &shard{m: map[string]string{}}
	}
	return s
}

// fnv1a is an inlined FNV-1a hash. We compute it ourselves instead of using
// hash/fnv to avoid allocating a hasher and to avoid the io.Writer path.
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

func (m *Map) shardFor(key string) *shard { return m.shards[fnv1a(key)&(numShards-1)] }

func (m *Map) Get(key string) (string, bool) {
	p := m.shardFor(key)
	p.mu.Lock()
	v, ok := p.m[key]
	p.mu.Unlock() // explicit, not defer (+8% here, see article)
	return v, ok
}

func (m *Map) Set(key, value string) {
	p := m.shardFor(key)
	p.mu.Lock()
	p.m[key] = value
	p.mu.Unlock()
}

func (m *Map) Delete(key string) {
	sh := m.shardFor(key)
	sh.mu.Lock()
	delete(sh.m, key)
	sh.mu.Unlock()
}

func (m *Map) Len() int {
	n := 0
	for _, sh := range m.shards {
		sh.mu.Lock()
		n += len(sh.m)
		sh.mu.Unlock()
	}
	return n
}
