package sharded

import "sync"

const (
	shards                = 2 << 7               // power of two, so we can mask instead of modulo
	fnvOffsetBasis uint64 = 14695981039346656037 // FNV-1a
	fnvPrime       uint64 = 1099511628211
)

type part struct {
	mu sync.Mutex
	m  map[string]string
}

type Map struct{ parts [shards]*part }

func (c *Map) at(key string) *part {
	h := fnvOffsetBasis
	for i := 0; i < len(key); i++ {
		h = (h ^ uint64(key[i])) * fnvPrime
	}
	return c.parts[h&(shards-1)]
}

func (c *Map) Get(key string) (string, bool) {
	p := c.at(key)
	p.mu.Lock()
	v, ok := p.m[key]
	p.mu.Unlock() // explicit, not defer (+8% here, see article)
	return v, ok
}

func (c *Map) Set(key, value string) {
	p := c.at(key)
	p.mu.Lock()
	p.m[key] = value
	p.mu.Unlock()
}
