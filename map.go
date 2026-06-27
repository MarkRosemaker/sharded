package sharded

import "sync"

const (
	// numShards is the number of independent map+lock partitions. It must be a
	// power of two so the shard index is a cheap bitmask of the key hash.
	numShards             = 2 << 7
	fnvOffsetBasis uint64 = 14695981039346656037 // FNV-1a
	fnvPrime       uint64 = 1099511628211
)

type part struct {
	mu sync.Mutex
	m  map[string]string
}

type Map struct{ parts [numShards]*part }

func NewMap() *Map {
	s := &Map{}
	for i := range s.parts {
		s.parts[i] = &part{m: make(map[string]string)}
	}
	return s
}

func (c *Map) at(key string) *part {
	h := fnvOffsetBasis
	for i := 0; i < len(key); i++ {
		h = (h ^ uint64(key[i])) * fnvPrime
	}
	return c.parts[h&(numShards-1)]
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

func (c *Map) Delete(key string) {
	sh := c.at(key)
	sh.mu.Lock()
	delete(sh.m, key)
	sh.mu.Unlock()
}

func (c *Map) Len() int {
	n := 0
	for _, sh := range c.parts {
		sh.mu.Lock()
		n += len(sh.m)
		sh.mu.Unlock()
	}
	return n
}
